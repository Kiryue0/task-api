// Composition root: tüm bağımlılıklar burada oluşturulup birbirine bağlanır.
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/Kiryue0/task-api/internal/adapter/driven/memory"
	"github.com/Kiryue0/task-api/internal/adapter/driven/postgres"
	httpadapter "github.com/Kiryue0/task-api/internal/adapter/driving/http"
	"github.com/Kiryue0/task-api/internal/core/port/driven"
	"github.com/Kiryue0/task-api/internal/core/service"
)

const listenAddr = ":8080"

func main() {
	if err := run(); err != nil {
		log.Fatalf("FATAL %v", err)
	}
}

func run() error {
	// SIGTERM: Kubernetes pod'u durdururken gönderir.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Yerel geliştirmede .env dosyası varsa yükle; ortamda tanımlı değerler önceliklidir.
	if n, err := loadDotEnv(".env"); err != nil {
		return err
	} else if n > 0 {
		log.Printf("INFO loaded %d variables from .env", n)
	}

	// --- Driven adapter: DB_* değişkenleri varsa PostgreSQL, hiç yoksa bellek ---
	dsn, dbConfigured, err := dsnFromEnv()
	if err != nil {
		return err
	}

	var (
		tableRepo driven.TableRepository
		taskRepo  driven.TaskRepository
	)
	if dbConfigured {
		db, closeDB, err := openPostgres(ctx, dsn)
		if err != nil {
			return err
		}
		defer closeDB()
		tableRepo = postgres.NewTableRepository(db)
		taskRepo = postgres.NewTaskRepository(db)
		log.Printf("INFO storage: postgres")
	} else {
		store := memory.NewStore()
		tableRepo, taskRepo = store, store
		log.Printf("WARN storage: in-memory (DB_* not set) - data is lost on restart and not shared between replicas")
	}

	// --- Core ---
	tableSvc := service.NewTableService(tableRepo)
	taskSvc := service.NewTaskService(taskRepo)

	// --- Driving adapter: HTTP ---
	// Handler'lar önekten habersizdir (/tables...); hepsi /api altına buradan bağlanır.
	api := http.NewServeMux()
	httpadapter.NewTableHandler(tableSvc).RegisterRoutes(api)
	httpadapter.NewTaskHandler(taskSvc).RegisterRoutes(api)

	mux := http.NewServeMux()
	mux.Handle("/api/", http.StripPrefix("/api", api))
	// /healthz bilerek /api dışında: Kubernetes probe'ları içindir, frontend'e ait değildir.
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{
		Addr:              listenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("INFO listening on %s", listenAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("http server: %w", err)
	case <-ctx.Done():
	}

	log.Printf("INFO shutting down")
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("http shutdown: %w", err)
	}
	return nil
}

// openPostgres, bağlantı havuzunu kurar, veritabanına ulaşılabildiğini doğrular ve
// şemayı hazırlar. Dönen kapatma fonksiyonu kaynakları serbest bırakır.
func openPostgres(ctx context.Context, dsn string) (*sql.DB, func(), error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("create pgx pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("ping database: %w", err)
	}

	// pgxpool'u database/sql arayüzü arkasında kullan.
	db := stdlib.OpenDBFromPool(pool)
	closeAll := func() {
		db.Close()
		pool.Close()
	}

	if err := postgres.EnsureSchema(ctx, db); err != nil {
		closeAll()
		return nil, nil, err
	}
	return db, closeAll, nil
}

// dbEnvKeys, veritabanı bağlantısı için gereken ortam değişkenleridir.
var dbEnvKeys = []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME"}

// dsnFromEnv, DB_* ortam değişkenlerinden bir PostgreSQL bağlantı URL'i üretir.
//   - Hiçbiri tanımlı değilse: configured=false döner (uygulama bellek-içi depoyla çalışır).
//   - Hepsi tanımlıysa: bağlantı URL'ini döner.
//   - Bir kısmı eksikse: hata döner. Yarım yapılandırma büyük ihtimalle bir hatadır
//     (örn. Secret'ta yanlış yazılmış bir anahtar); sessizce belleğe düşüp veri
//     kaybettirmek yerine açılışta durulur.
func dsnFromEnv() (dsn string, configured bool, err error) {
	vals := make(map[string]string, len(dbEnvKeys))
	var missing []string
	for _, k := range dbEnvKeys {
		v := os.Getenv(k)
		if v == "" {
			missing = append(missing, k)
		}
		vals[k] = v
	}
	if len(missing) == len(dbEnvKeys) {
		return "", false, nil
	}
	if len(missing) > 0 {
		return "", false, fmt.Errorf("incomplete database configuration, missing: %v (set all DB_* variables, or none to use in-memory storage)", missing)
	}

	// url.URL kullanmak, şifredeki özel karakterlerin doğru escape edilmesini sağlar.
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(vals["DB_USER"], vals["DB_PASSWORD"]),
		Host:   net.JoinHostPort(vals["DB_HOST"], vals["DB_PORT"]),
		Path:   "/" + vals["DB_NAME"],
	}
	return u.String(), true, nil
}

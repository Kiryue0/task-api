// Package postgres, driven port'ların PostgreSQL adapter'ıdır.
// database/sql üzerinden, pgx driver'ı (pgxpool ile) kullanılarak çalışır.
package postgres

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// userSchema, kullanıcı tablolarının oluşturulduğu Postgres şemasıdır (schema.sql ile aynı).
const userSchema = "user_tables"

// Postgres hata kodları: https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	codeUndefinedTable  = "42P01"
	codeDuplicateTable  = "42P07"
	codeUniqueViolation = "23505"
)

// schemaSQL, derleme anında binary'nin içine gömülür; çalışma zamanında dosya gerekmez.
//
//go:embed schema.sql
var schemaSQL string

// schemaLockID, şema kurulumunu sıraya sokan advisory lock'un anahtarıdır (keyfi bir sabit).
const schemaLockID = 7_340_001

// EnsureSchema, schema.sql'i çalıştırır. Uygulama açılışında bir kez çağrılır;
// tekrar çalıştırmak güvenlidir.
//
// Birden çok replika aynı anda açılırsa "CREATE ... IF NOT EXISTS" bile çakışıp
// unique violation verebilir. Bu yüzden kurulum bir transaction içinde advisory lock
// alınarak yapılır: replikalar sırayla girer, kilit commit/rollback ile kendiliğinden bırakılır.
func EnsureSchema(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("ensure schema: begin: %w", err)
	}
	defer tx.Rollback() // commit başarılıysa etkisizdir

	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, schemaLockID); err != nil {
		return fmt.Errorf("ensure schema: lock: %w", err)
	}
	if _, err := tx.ExecContext(ctx, schemaSQL); err != nil {
		return fmt.Errorf("ensure schema: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("ensure schema: commit: %w", err)
	}
	return nil
}

// qualifiedTable, tablo adını şemayla birlikte güvenli şekilde tırnaklar:
// alisveris -> "user_tables"."alisveris". Ad zaten core'da doğrulanmış olsa da
// SQL'e hiçbir zaman tırnaksız girmez (ikinci savunma hattı).
func qualifiedTable(name string) string {
	return pgx.Identifier{userSchema, name}.Sanitize()
}

// pgErrorCode, hata bir Postgres hatasıysa kodunu, değilse boş string döner.
func pgErrorCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

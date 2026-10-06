# task-api

Go standart kütüphanesi (`net/http`) ve PostgreSQL ile yazılmış minimal bir REST API.
API üzerinden **tablo oluşturulur**. Her tablo Postgres'te gerçek, ayrı bir tablodur.
Her tablonun içinde görevler (`id`, `title`) tutulur ve tam CRUD yapılır.
Hexagonal (Ports & Adapters) mimari kullanır.

## Gereksinimler

- Go 1.25+ (pgx v5.11 bunu istiyor; daha eski bir `go` komutu uygun toolchain'i otomatik indirir)
- PostgreSQL **isteğe bağlı**. Yoksa uygulama bellek-içi depoyla çalışır (aşağıya bak).
  Varsa DB kullanıcısının **şema ve tablo oluşturma yetkisi** (`CREATE`) olmalı.

## Ortam değişkenleri

Bu değişkenler depolama modunu belirler:

| Durum | Sonuç |
|---|---|
| **Hiçbiri** tanımlı değil | **Bellek modu**: veriler sürecin belleğinde tutulur. Açılışta `WARN storage: in-memory` loglanır |
| **Hepsi** tanımlı | **Postgres modu**: gerçek veritabanına bağlanır. Açılışta `INFO storage: postgres` loglanır |
| **Bir kısmı** tanımlı | Uygulama hata verip kapanır ve eksikleri listeler. Yarım yapılandırma büyük ihtimalle bir yazım hatasıdır; sessizce belleğe düşüp veri kaybettirmez |

**Bellek modunun sınırları:** Pod yeniden başlayınca veriler kaybolur. Her replika kendi
ayrı verisini tutar, bu yüzden bu modda **tek replika** (`replicas: 1`) çalıştır. Postgres'e
geçince bellekteki veriler taşınmaz.

| Değişken      | Açıklama            | Örnek       |
|---------------|---------------------|-------------|
| `DB_HOST`     | Postgres host'u     | `localhost` |
| `DB_PORT`     | Postgres portu      | `5432`      |
| `DB_USER`     | Kullanıcı adı       | `app`       |
| `DB_PASSWORD` | Şifre               | `secret`    |
| `DB_NAME`     | Veritabanı adı      | `tasks`     |

HTTP sunucusu `:8080` portunu dinler.

## Çalıştırma

```bash
# (İsteğe bağlı) Docker ile yerel bir Postgres:
docker run -d --name task-api-postgres -e POSTGRES_USER=app -e POSTGRES_PASSWORD=secret \
  -e POSTGRES_DB=tasks -p 5432:5432 postgres:16-alpine

cp .env.example .env     # değerleri gerekirse düzenle
go run ./cmd/server
```

### `.env` dosyası

Uygulama açılışta çalıştığı klasördeki `.env` dosyasını otomatik okur (ek kütüphane yok).
- Dosya yoksa sessizce geçer. Kubernetes'te dosya olmaz, değerler ConfigMap ve Secret'tan gelir.
- Ortamda zaten tanımlı bir değişkenin **üzerine yazmaz**. Örneğin `DB_PORT=5433 go run ./cmd/server`
  dersen `.env`'deki port yerine 5433 kullanılır.
- `.env` git'e (`.gitignore`) ve Docker image'ına (`.dockerignore`) girmez. Repoda yalnızca
  şifresiz şablon `.env.example` durur.

Testler (Postgres gerekmez; core servisler bellek-içi adapter ile test edilir):

```bash
go test ./...
```

## Endpoint'ler

Tüm uygulama endpoint'leri `/api` önekiyle başlar. `/healthz` bunun dışındadır.

### Tablolar

| Metot    | Yol               | Gövde              | Başarılı      | Hatalar                          |
|----------|-------------------|--------------------|---------------|----------------------------------|
| `POST`   | `/api/tables`         | `{"name":"okul"}`  | `201`         | `400` geçersiz ad, `409` zaten var |
| `GET`    | `/api/tables`         | –                  | `200` liste   |                                  |
| `DELETE` | `/api/tables/{table}` | –                  | `204`         | `404` tablo yok                  |

> `DELETE /api/tables/{table}` tabloyu **içindeki tüm verilerle birlikte** kalıcı olarak siler.

**Tablo adı kuralları:** harfle başlamalı; yalnızca `a-z`, `0-9` ve `_` içerebilir; en fazla
63 karakter olabilir. Büyük harfler küçüğe çevrilir (`Okul` → `okul`). Türkçe karakterler
(`ş`, `ı`...) kabul edilmez.

### Bir tablodaki görevler

| Metot    | Yol                          | Gövde              | Başarılı      | Hatalar                              |
|----------|------------------------------|--------------------|---------------|--------------------------------------|
| `POST`   | `/api/tables/{table}/tasks`      | `{"title":"..."}`  | `201` + görev | `400` boş title, `404` tablo yok     |
| `GET`    | `/api/tables/{table}/tasks`      | –                  | `200` liste   | `404` tablo yok                      |
| `GET`    | `/api/tables/{table}/tasks/{id}` | –                  | `200` + görev | `400` geçersiz id, `404`             |
| `PUT`    | `/api/tables/{table}/tasks/{id}` | `{"title":"..."}`  | `200` + görev | `400`, `404`                         |
| `DELETE` | `/api/tables/{table}/tasks/{id}` | –                  | `204`         | `400`, `404`                         |

### Diğer

| `GET` | `/healthz` | Kubernetes probe'ları için; her zaman `200` |
|-------|------------|---------------------------------------------|

Hatalar `{"error":"..."}` olarak döner.

### Örnek akış

```bash
curl -X POST localhost:8080/api/tables -d '{"name":"alisveris"}'
curl -X POST localhost:8080/api/tables/alisveris/tasks -d '{"title":"süt al"}'
curl localhost:8080/api/tables/alisveris/tasks
curl -X PUT localhost:8080/api/tables/alisveris/tasks/1 -d '{"title":"yarım yağlı süt al"}'
curl -X DELETE localhost:8080/api/tables/alisveris/tasks/1
curl -X DELETE localhost:8080/api/tables/alisveris
```

## Veritabanı yapısı

- Açılışta `internal/adapter/driven/postgres/schema.sql` çalıştırılır ve `user_tables`
  şeması oluşturulur (`go:embed` ile binary'ye gömülü; tekrar çalıştırmak güvenlidir).
- API ile oluşturulan her tablo bu şemada gerçek bir Postgres tablosudur:

  ```sql
  CREATE TABLE "user_tables"."alisveris" (
      id    SERIAL PRIMARY KEY,
      title TEXT   NOT NULL CHECK (btrim(title) <> '')
  );
  ```

- psql ile görmek için: `\dt user_tables.*`

**Güvenlik (SQL injection):** Tablo adı SQL'e parametre olarak verilemez. Bu yüzden iki katlı
koruma var:
1. Core, adı sıkı bir kurala göre doğrular (`domain.ParseTableName`).
2. Postgres adapter adı her zaman `pgx.Identifier` ile tırnaklar.

## Mimari

```
cmd/server/main.go                                  -- composition root (wiring)
internal/core/domain/{task,table}.go                -- varlıklar + domain hataları + ad kuralı
internal/core/port/driving/{task,table}_service.go  -- driving port'lar
internal/core/port/driven/{task,table}_repository.go -- driven port'lar
internal/core/service/{task,table}_service.go       -- core: iş mantığı
internal/adapter/driving/http/                      -- HTTP adapter
internal/adapter/driven/postgres/                   -- Postgres adapter + schema.sql
internal/adapter/driven/memory/                     -- bellek-içi adapter (DB yokken)
```

Bağımlılık yönü: adapter'lar → port'lar ← core. `internal/core` altındaki hiçbir paket
adapter'ları import etmez. Bunu kontrol etmek için:

```bash
go list -f '{{.ImportPath}}: {{join .Imports " "}}' ./internal/core/...
```

## Docker

```bash
docker build -t task-api:1.0 .
```

Çok aşamalı (multi-stage) build:
- İlk aşama testleri çalıştırır (`go test ./...`) ve binary'yi derler. Test kırılırsa build durur.
- Son image `distroless/static` üzerindedir, root olmayan kullanıcıyla (uid 65532) çalışır (~13 MB).

Build sırasında internet gerekir: Go modülleri, `golang:1.25-alpine` ve `gcr.io/distroless/static-debian12` indirilir.

## Kubernetes'e deploy için notlar

Manifest yazarken bilmen gerekenler:

| Konu | Değer |
|---|---|
| Container portu | `8080` (sabit) |
| Ortam değişkenleri | `DB_HOST`, `DB_PORT`, `DB_NAME` (ConfigMap); `DB_USER`, `DB_PASSWORD` (Secret). Hiçbiri verilmezse bellek modu |
| Probe | `GET /healthz` → `200` (DB'yi kontrol etmez, sadece "süreç ayakta" der) |
| Kullanıcı | uid/gid `65532`; `runAsNonRoot: true` ve `readOnlyRootFilesystem: true` ile uyumlu |
| Kapanış | SIGTERM'de yeni istek almayı bırakır, en fazla 10 sn bekler (`terminationGracePeriodSeconds` > 10 olmalı) |
| DB yetkisi | Kullanıcı şema ve tablo oluşturabilmeli (`CREATE` yetkisi) |
| Replika | Bellek modunda **1**. Postgres modunda birden çok replika güvenli (açılıştaki şema kurulumu advisory lock ile sıraya alınır) |
| Açılış sırası | Postgres hazır değilse uygulama ~5 sn sonra çıkar, Kubernetes yeniden başlatır (kendiliğinden düzelir) |

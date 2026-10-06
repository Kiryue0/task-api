# ---- 1. aşama: derleme ----
# Go derleyicisi yalnızca bu aşamada var; son image'a girmez.
FROM golang:1.25-alpine AS build
WORKDIR /src

# Önce sadece bağımlılık listesini kopyala: kod değişse bile go.mod/go.sum
# değişmedikçe bu katman cache'ten gelir, bağımlılıklar tekrar indirilmez.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Testler image build'inin parçası: biri kırılırsa build durur ve image hiç üretilmez.
# Böylece CI agent'ında Go kurulu olmasına gerek kalmaz. (Testler Postgres istemez.)
RUN CGO_ENABLED=0 go test ./...
# CGO_ENABLED=0: C kütüphanelerine bağımlı olmayan, tek başına çalışan bir binary üretir.
# -trimpath ve -ldflags="-s -w": binary'den yerel yolları ve debug bilgisini atarak küçültür.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/task-api ./cmd/server

# ---- 2. aşama: çalışma ----
# distroless/static: shell, paket yöneticisi vb. yok; sadece binary'nin ihtiyaç duyduğu
# asgari dosyalar (CA sertifikaları, saat dilimi verisi) var.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/task-api /task-api
EXPOSE 8080
# Kullanıcı isimle değil numarayla verilir: Kubernetes'te "runAsNonRoot: true" açıkken
# kubelet isimli bir kullanıcının (örn. "nonroot") root olmadığını doğrulayamaz ve pod'u başlatmaz.
# 65532, distroless'ın "nonroot" kullanıcısının uid/gid'idir.
USER 65532:65532
ENTRYPOINT ["/task-api"]

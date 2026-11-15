# Ship

Layanan pelacakan paket. Backend Go untuk REST, konkurensi, dan observability, dengan satu binary tanpa runtime tambahan.

## Tech stack

| Bagian | Teknologi |
|---|---|
| Bahasa | Go 1.24+ |
| HTTP | `net/http` (router bawaan, pola `GET /path`) |
| Frontend | HTML + CSS inline, disajikan dari binary |
| Rencana | PostgreSQL (pgx), Redis, Kafka/queue, Prometheus `client_golang` |

## Alur singkat

1. Pengirim membuat resi, tersimpan di database.
2. Kurir memindai resi di tiap titik, status masuk ke tabel riwayat.
3. Webhook dikirim ke sistem pengirim lewat antrean.
4. Publik membuka `/` lalu memasukkan nomor resi untuk melihat timeline.
5. Operator memantau metrik di `/metrics`.

## Menjalankan

```bash
go run .
```

Server di `http://localhost:8081`, health check di `/health`.

```bash
go build -o ship .
go vet ./...
go test ./...
```

Docker:

```bash
docker build -t ship-go .
docker run --rm -p 8081:8081 ship-go
```

## Struktur

```
main.go        entrypoint, router, halaman, handler API
store.go       penyimpanan paket dan riwayat di memori
main_test.go   uji alur buat, scan, lacak
Dockerfile     build multi-stage, runtime distroless nonroot
go.mod         modul github.com/izanmaizan/ship-go
```

## Status

API jalan: `POST /api/paket`, `POST /api/paket/{resi}/scan`, `GET /api/paket/{resi}`. Penyimpanan masih di memori, belum ada database, antrean, atau metrik.

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
```

## Struktur

```
main.go        entrypoint, router, halaman
go.mod         modul github.com/izanmaizan/ship-go
```

## Status

Scaffold. Endpoint `/health` dan halaman lacak sudah jalan. Belum ada database, cache, antrean, atau metrik.

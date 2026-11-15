package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	app          = "ship-go"
	maxBodyBytes = 1 << 14
	maxResi      = 50
	maxTujuan    = 200
	maxStatus    = 50
	maxLokasi    = 100
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	log.Printf("%s listen :%s", app, port)
	log.Fatal(http.ListenAndServe(":"+port, newMux(NewStore())))
}

func newMux(store *Store) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "app": app})
	})

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(indexHTML))
	})

	mux.HandleFunc("POST /api/paket", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Resi   string `json:"resi"`
			Tujuan string `json:"tujuan"`
		}
		if !readJSON(w, r, &in) {
			return
		}
		in.Resi, in.Tujuan = strings.TrimSpace(in.Resi), strings.TrimSpace(in.Tujuan)
		if !panjangValid(in.Resi, maxResi) || !panjangValid(in.Tujuan, maxTujuan) {
			writeError(w, http.StatusBadRequest, "resi atau tujuan kosong / terlalu panjang")
			return
		}
		paket, err := store.Buat(in.Resi, in.Tujuan)
		if errors.Is(err, ErrSudahAda) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, paket)
	})

	mux.HandleFunc("POST /api/paket/{resi}/scan", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Status string `json:"status"`
			Lokasi string `json:"lokasi"`
		}
		if !readJSON(w, r, &in) {
			return
		}
		in.Status, in.Lokasi = strings.TrimSpace(in.Status), strings.TrimSpace(in.Lokasi)
		if !panjangValid(in.Status, maxStatus) || len(in.Lokasi) > maxLokasi {
			writeError(w, http.StatusBadRequest, "status kosong atau lokasi terlalu panjang")
			return
		}
		paket, err := store.Catat(r.PathValue("resi"), Langkah{Status: in.Status, Lokasi: in.Lokasi})
		if errors.Is(err, ErrTidakKetemu) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, paket)
	})

	mux.HandleFunc("GET /api/paket/{resi}", func(w http.ResponseWriter, r *http.Request) {
		paket, err := store.Lacak(r.PathValue("resi"))
		if errors.Is(err, ErrTidakKetemu) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, paket)
	})

	return mux
}

func readJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, "JSON tidak valid")
		return false
	}
	return true
}

func panjangValid(s string, maks int) bool {
	return len(s) > 0 && len(s) <= maks
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

const indexHTML = `<!doctype html>
<html lang="id" data-theme="light"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Ship - Lacak Paket</title>
<link href="https://cdn.jsdelivr.net/npm/daisyui@5" rel="stylesheet" type="text/css">
<script src="https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4"></script>
</head><body class="min-h-screen flex flex-col bg-base-100">
<header class="navbar border-b border-base-200 px-4 sm:px-8">
<div class="flex-1"><a href="/" class="text-lg font-semibold">Ship</a></div>
<nav class="flex-none"><a href="#bantuan" class="btn btn-ghost btn-sm">Bantuan</a></nav>
</header>
<main class="flex-1 w-full max-w-2xl mx-auto px-4 py-10 sm:py-16">
<section>
<h1 class="text-3xl font-semibold">Lacak paket</h1>
<p class="text-base-content/70 mt-2">Masukkan nomor resi untuk melihat riwayat pengiriman.</p>
<form id="lacak" class="flex flex-col sm:flex-row gap-2 mt-6">
<input id="resi" class="input input-bordered w-full" placeholder="Contoh: SHP1234" aria-label="Nomor resi" maxlength="50" required>
<button class="btn btn-primary sm:w-32">Lacak</button>
</form>
</section>
<section id="hasil" class="mt-8" aria-live="polite"></section>
<section id="bantuan" class="mt-16">
<h2 class="text-lg font-semibold mb-2">Pertanyaan umum</h2>
<div class="collapse collapse-arrow border border-base-300 bg-base-100">
<input type="radio" name="faq" checked>
<div class="collapse-title font-medium">Apa itu nomor resi?</div>
<div class="collapse-content text-sm text-base-content/70"><p>Nomor unik dari pengirim untuk setiap paket. Tulis persis seperti di label atau bukti kirim.</p></div>
</div>
<div class="collapse collapse-arrow border border-base-300 bg-base-100 mt-2">
<input type="radio" name="faq">
<div class="collapse-title font-medium">Kenapa riwayat kosong?</div>
<div class="collapse-content text-sm text-base-content/70"><p>Paket baru terdaftar dan belum dipindai di titik mana pun. Cek lagi beberapa saat.</p></div>
</div>
<div class="collapse collapse-arrow border border-base-300 bg-base-100 mt-2">
<input type="radio" name="faq">
<div class="collapse-title font-medium">Kenapa nomor resi tidak ditemukan?</div>
<div class="collapse-content text-sm text-base-content/70"><p>Periksa penulisan. Huruf dan angka harus sama persis, termasuk huruf besar.</p></div>
</div>
</section>
</main>
<footer class="footer footer-center p-6 text-sm text-base-content/60 border-t border-base-200">
<p>Ship, layanan pelacakan paket</p>
</footer>
<script>
const hasil = document.getElementById("hasil");
function el(tag, cls, teks) {
  const node = document.createElement(tag);
  if (cls) node.className = cls;
  if (teks) node.textContent = teks;
  return node;
}
function tampilkanError(pesan) {
  const box = el("div", "alert alert-error", pesan);
  box.setAttribute("role", "alert");
  hasil.replaceChildren(box);
}
function tampilkanRiwayat(data) {
  const kepala = el("div", "flex items-start justify-between gap-4 mb-4");
  const judul = el("div");
  judul.append(el("p", "text-sm text-base-content/60", "Nomor resi"), el("p", "font-semibold font-mono", data.resi));
  const tujuan = el("div", "text-right");
  tujuan.append(el("p", "text-sm text-base-content/60", "Tujuan"), el("p", "font-semibold", data.tujuan));
  kepala.append(judul, tujuan);
  const wadah = el("div", "card border border-base-300 bg-base-100");
  const isi = el("div", "card-body p-5");
  isi.append(kepala);
  if (!data.riwayat.length) {
    isi.append(el("p", "text-base-content/70", "Belum ada riwayat pengiriman."));
    wadah.append(isi);
    hasil.replaceChildren(wadah);
    return;
  }
  const daftar = data.riwayat.slice().reverse();
  const ul = el("ul", "timeline timeline-vertical timeline-compact");
  daftar.forEach((langkah, i) => {
    const li = el("li");
    const waktu = el("div", "timeline-start text-xs text-base-content/60 text-right pe-3 self-center", new Date(langkah.waktu).toLocaleString("id-ID"));
    const tengah = el("div", "timeline-middle");
    tengah.append(el("span", i === 0 ? "status status-primary" : "status"));
    const akhir = el("div", "timeline-end pb-6");
    const baris = el("div", "flex items-center gap-2");
    baris.append(el("span", "font-medium", langkah.status));
    if (i === 0) baris.append(el("span", "badge badge-primary badge-sm", "Terbaru"));
    akhir.append(baris);
    if (langkah.lokasi) akhir.append(el("p", "text-sm text-base-content/70", langkah.lokasi));
    li.append(waktu, tengah, akhir);
    if (i < daftar.length - 1) li.append(el("hr"));
    ul.append(li);
  });
  isi.append(ul);
  wadah.append(isi);
  hasil.replaceChildren(wadah);
}
document.getElementById("lacak").addEventListener("submit", async (e) => {
  e.preventDefault();
  const resi = document.getElementById("resi").value.trim();
  hasil.replaceChildren(el("p", "text-base-content/70", "Mencari..."));
  const res = await fetch("/api/paket/" + encodeURIComponent(resi));
  const data = await res.json();
  if (!res.ok) {
    tampilkanError(data.error);
    return;
  }
  tampilkanRiwayat(data);
});
</script></body></html>`

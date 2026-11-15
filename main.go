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
</head><body class="min-h-screen bg-base-200 flex justify-center p-6">
<main class="card bg-base-100 shadow w-full max-w-lg mt-12"><div class="card-body">
<h1 class="card-title text-2xl">Ship</h1>
<p class="text-sm text-base-content/70">Lacak paket dengan nomor resi.</p>
<form id="lacak" class="join w-full mt-4"><input id="resi" class="input join-item w-full" placeholder="Nomor resi" aria-label="Nomor resi" maxlength="50" required><button class="btn btn-primary join-item">Lacak</button></form>
<div id="hasil" class="mt-6"></div>
</div></main>
<script>
const hasil = document.getElementById("hasil");
function el(tag, cls, teks) {
  const node = document.createElement(tag);
  if (cls) node.className = cls;
  if (teks) node.textContent = teks;
  return node;
}
function tampilkanError(pesan) {
  const box = el("div", "alert alert-error alert-soft", pesan);
  box.setAttribute("role", "alert");
  hasil.append(box);
}
document.getElementById("lacak").addEventListener("submit", async (e) => {
  e.preventDefault();
  hasil.replaceChildren();
  const resi = document.getElementById("resi").value.trim();
  const res = await fetch("/api/paket/" + encodeURIComponent(resi));
  const data = await res.json();
  if (!res.ok) {
    tampilkanError(data.error);
    return;
  }
  if (!data.riwayat.length) {
    hasil.append(el("p", "text-sm text-base-content/70", "Belum ada riwayat untuk " + data.resi + "."));
    return;
  }
  const daftar = data.riwayat.slice().reverse();
  const ul = el("ul", "timeline timeline-vertical timeline-compact");
  daftar.forEach((langkah, i) => {
    const li = el("li");
    const tengah = el("div", "timeline-middle");
    tengah.append(el("span", "status status-primary"));
    const akhir = el("div", "timeline-end mb-6");
    akhir.append(el("div", "font-medium", langkah.status));
    const meta = [langkah.lokasi, new Date(langkah.waktu).toLocaleString("id-ID")].filter(Boolean).join(" · ");
    akhir.append(el("div", "text-xs text-base-content/70", meta));
    li.append(tengah, akhir);
    if (i < daftar.length - 1) li.append(el("hr", "bg-base-300"));
    ul.append(li);
  });
  hasil.append(ul);
});
</script></body></html>`

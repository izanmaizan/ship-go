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
<html lang="id"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Ship - Lacak Paket</title>
<style>
*{box-sizing:border-box;margin:0}body{font-family:ui-sans-serif,system-ui;background:#08090c;color:#e6e8ef;min-height:100vh;display:grid;place-items:center;padding:24px 0}
.card{width:min(560px,92vw);background:linear-gradient(160deg,#11131a,#0b0c10);border:1px solid #1d2130;border-radius:20px;padding:32px}
h1{font-size:22px;letter-spacing:-.02em}p{color:#8b93a7;font-size:14px;margin-top:8px}
form{margin-top:24px;display:flex;gap:8px}input{flex:1;background:#0d0f15;border:1px solid #232838;color:#e6e8ef;padding:12px 14px;border-radius:12px;font-size:14px}
button{background:#3b82f6;border:0;color:#fff;padding:12px 18px;border-radius:12px;font-weight:600;cursor:pointer}
ol{list-style:none;margin-top:24px;border-left:1px solid #232838;padding-left:16px}li{padding:8px 0;font-size:14px}li small{display:block;color:#8b93a7}
.err{color:#f87171;margin-top:16px;font-size:14px}
</style></head><body><div class="card"><h1>Ship</h1><p>Lacak paket dengan nomor resi.</p>
<form id="lacak"><input id="resi" placeholder="Nomor resi" aria-label="Nomor resi" maxlength="50" required><button>Lacak</button></form>
<div id="hasil"></div></div>
<script>
const hasil = document.getElementById("hasil");
document.getElementById("lacak").addEventListener("submit", async (e) => {
  e.preventDefault();
  hasil.replaceChildren();
  const resi = document.getElementById("resi").value.trim();
  const res = await fetch("/api/paket/" + encodeURIComponent(resi));
  const data = await res.json();
  if (!res.ok) {
    const p = document.createElement("p");
    p.className = "err";
    p.textContent = data.error;
    hasil.append(p);
    return;
  }
  const ol = document.createElement("ol");
  for (const langkah of data.riwayat.slice().reverse()) {
    const li = document.createElement("li");
    li.textContent = langkah.status;
    const small = document.createElement("small");
    small.textContent = [langkah.lokasi, new Date(langkah.waktu).toLocaleString("id-ID")].filter(Boolean).join(" · ");
    li.append(small);
    ol.append(li);
  }
  if (!data.riwayat.length) {
    const p = document.createElement("p");
    p.textContent = "Belum ada riwayat untuk " + data.resi + ".";
    hasil.append(p);
    return;
  }
  hasil.append(ol);
});
</script></body></html>`

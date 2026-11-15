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
	store := NewStore()
	if os.Getenv("DUMMY") != "0" {
		SeedDummy(store)
		log.Printf("data dummy aktif (matikan dengan DUMMY=0)")
	}
	log.Printf("%s listen :%s", app, port)
	log.Fatal(http.ListenAndServe(":"+port, newMux(store)))
}

func newMux(store *Store) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "app": app})
	})

	mux.HandleFunc("GET /api/paket", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, store.Daftar())
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

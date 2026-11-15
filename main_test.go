package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAlurLacak(t *testing.T) {
	h := newMux(NewStore())
	do := func(method, path, body string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(method, path, bytes.NewBufferString(body)))
		return rec
	}
	cek := func(nama string, got, want int) {
		if got != want {
			t.Fatalf("%s: status %d, mau %d", nama, got, want)
		}
	}

	cek("buat", do("POST", "/api/paket", `{"resi":"SHP1","tujuan":"Solok"}`).Code, http.StatusCreated)
	cek("duplikat", do("POST", "/api/paket", `{"resi":"SHP1","tujuan":"Solok"}`).Code, http.StatusConflict)
	cek("scan", do("POST", "/api/paket/SHP1/scan", `{"status":"Dikirim","lokasi":"Gudang"}`).Code, http.StatusOK)
	cek("scan resi tak ada", do("POST", "/api/paket/XX/scan", `{"status":"Dikirim"}`).Code, http.StatusNotFound)
	cek("lacak tak ada", do("GET", "/api/paket/XX", "").Code, http.StatusNotFound)
	cek("resi kosong", do("POST", "/api/paket", `{"resi":"  ","tujuan":"x"}`).Code, http.StatusBadRequest)
	cek("field asing", do("POST", "/api/paket", `{"resi":"a","tujuan":"b","x":1}`).Code, http.StatusBadRequest)
	cek("beranda", do("GET", "/", "").Code, http.StatusOK)
	cek("health", do("GET", "/health", "").Code, http.StatusOK)

	var p Paket
	if err := json.Unmarshal(do("GET", "/api/paket/SHP1", "").Body.Bytes(), &p); err != nil {
		t.Fatal(err)
	}
	if len(p.Riwayat) != 1 || p.Riwayat[0].Status != "Dikirim" || p.Riwayat[0].Waktu.IsZero() {
		t.Fatalf("riwayat salah: %+v", p.Riwayat)
	}
}

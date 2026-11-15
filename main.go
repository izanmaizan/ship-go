package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

type health struct {
	Status string `json:"status"`
	App    string `json:"app"`
}

func main() {
	app := "ship-go"
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(health{Status: "ok", App: app})
	})
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(indexHTML))
	})
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	log.Printf("%s listen :%s", app, port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

const indexHTML = `<!doctype html>
<html lang="id"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Ship - Lacak Paket</title>
<style>
*{box-sizing:border-box;margin:0}body{font-family:ui-sans-serif,system-ui;background:#08090c;color:#e6e8ef;min-height:100vh;display:grid;place-items:center}
.card{width:min(560px,92vw);background:linear-gradient(160deg,#11131a,#0b0c10);border:1px solid #1d2130;border-radius:20px;padding:32px}
h1{font-size:22px;letter-spacing:-.02em}p{color:#8b93a7;font-size:14px;margin-top:8px}
form{margin-top:24px;display:flex;gap:8px}input{flex:1;background:#0d0f15;border:1px solid #232838;color:#e6e8ef;padding:12px 14px;border-radius:12px;font-size:14px}
button{background:#3b82f6;border:0;color:#fff;padding:12px 18px;border-radius:12px;font-weight:600;cursor:pointer}
</style></head><body><div class="card"><h1>Ship</h1><p>Lacak paket dengan nomor resi. Scaffold Go siap dikembangkan.</p>
<form onsubmit="event.preventDefault();alert('Belum terhubung API')"><input placeholder="Nomor resi" aria-label="Nomor resi"><button>Lacak</button></form></div></body></html>`

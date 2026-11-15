package main

const indexHTML = `<!doctype html>
<html lang="id" data-theme="light"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Ship - Lacak Paket</title>
<link href="https://cdn.jsdelivr.net/npm/daisyui@5" rel="stylesheet" type="text/css">
<script src="https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4"></script>
</head><body class="min-h-screen flex flex-col bg-base-200">
<header class="navbar bg-base-100 border-b border-base-300 px-4 sm:px-8">
<div class="flex-1 gap-2"><span class="font-semibold text-lg">Ship</span><span class="badge badge-ghost badge-sm">lacak</span></div>
<div class="flex-none"><a href="#contoh" class="btn btn-ghost btn-sm">Contoh resi</a></div>
</header>
<main class="flex-1 w-full max-w-3xl mx-auto px-4 py-10 space-y-8">
<section class="card bg-base-100 border border-base-300">
<div class="card-body">
<h1 class="card-title text-2xl">Lacak paket</h1>
<p class="text-base-content/70 text-sm">Masukkan nomor resi untuk melihat riwayat pengiriman.</p>
<form id="lacak" class="join w-full mt-2">
<input id="resi" class="input join-item w-full font-mono uppercase" placeholder="SHP1001" aria-label="Nomor resi" maxlength="50" autocomplete="off" required>
<button class="btn btn-primary join-item">Lacak</button>
</form>
</div>
</section>
<section id="hasil" aria-live="polite"></section>
<section id="contoh" class="card bg-base-100 border border-base-300">
<div class="card-body">
<h2 class="card-title text-base">Contoh resi</h2>
<p class="text-sm text-base-content/70">Data dummy untuk dicek. Klik baris untuk melacak.</p>
<div class="overflow-x-auto mt-2">
<table class="table">
<thead><tr><th>Resi</th><th>Tujuan</th><th>Status terakhir</th></tr></thead>
<tbody id="daftar"></tbody>
</table>
</div>
</div>
</section>
</main>
<footer class="footer footer-center p-6 text-sm text-base-content/60">
<p>Ship. Data di halaman ini hanya contoh.</p>
</footer>
<script>
const hasil = document.getElementById("hasil");
const daftar = document.getElementById("daftar");
const input = document.getElementById("resi");

function el(tag, cls, teks) {
  const node = document.createElement(tag);
  if (cls) node.className = cls;
  if (teks !== undefined) node.textContent = teks;
  return node;
}

function waktu(iso) {
  return new Date(iso).toLocaleString("id-ID", { dateStyle: "medium", timeStyle: "short" });
}

function tampilkanError(pesan) {
  const box = el("div", "alert alert-error", pesan);
  box.setAttribute("role", "alert");
  hasil.replaceChildren(box);
}

function tampilkanPaket(data) {
  const terbaru = data.riwayat.length ? data.riwayat[data.riwayat.length - 1] : null;
  const card = el("article", "card bg-base-100 border border-base-300");
  const body = el("div", "card-body");

  const head = el("div", "flex flex-wrap items-start justify-between gap-4");
  const kiri = el("div");
  kiri.append(el("p", "text-xs uppercase tracking-wide text-base-content/60", "Nomor resi"), el("p", "font-mono font-semibold text-lg", data.resi));
  const kanan = el("div", "text-right");
  kanan.append(el("p", "text-xs uppercase tracking-wide text-base-content/60", "Tujuan"), el("p", "font-semibold", data.tujuan));
  head.append(kiri, kanan);
  body.append(head);

  if (terbaru) {
    const ringkas = el("div", "alert alert-soft mt-4");
    ringkas.append(el("span", "badge badge-primary", "Terbaru"), el("span", "font-medium", terbaru.status));
    ringkas.append(el("span", "text-sm text-base-content/70", waktu(terbaru.waktu)));
    body.append(ringkas);
  }

  if (!data.riwayat.length) {
    body.append(el("p", "text-base-content/70 mt-4", "Belum ada riwayat pengiriman."));
  } else {
    const urut = data.riwayat.slice().reverse();
    const ul = el("ul", "timeline timeline-vertical timeline-compact mt-4");
    urut.forEach((langkah, i) => {
      const li = el("li");
      if (i > 0) li.append(el("hr", "bg-primary"));
      li.append(el("div", "timeline-start text-xs text-base-content/60 self-center pe-3 text-right", waktu(langkah.waktu)));
      const tengah = el("div", "timeline-middle");
      tengah.append(el("span", i === 0 ? "status status-primary" : "status"));
      const akhir = el("div", "timeline-end pb-6");
      akhir.append(el("p", i === 0 ? "font-semibold" : "font-medium", langkah.status));
      if (langkah.lokasi) akhir.append(el("p", "text-sm text-base-content/70", langkah.lokasi));
      li.append(tengah, akhir);
      if (i < urut.length - 1) li.append(el("hr", "bg-base-300"));
      ul.append(li);
    });
    body.append(ul);
  }

  card.append(body);
  hasil.replaceChildren(card);
}

async function lacak(resi) {
  input.value = resi;
  hasil.replaceChildren(el("p", "text-sm text-base-content/70 px-1", "Mencari " + resi + "..."));
  const res = await fetch("/api/paket/" + encodeURIComponent(resi));
  const data = await res.json();
  if (!res.ok) {
    tampilkanError(data.error);
    return;
  }
  tampilkanPaket(data);
  hasil.scrollIntoView({ behavior: "smooth", block: "start" });
}

document.getElementById("lacak").addEventListener("submit", (e) => {
  e.preventDefault();
  lacak(input.value.trim());
});

async function muatDaftar() {
  const res = await fetch("/api/paket");
  const rows = await res.json();
  daftar.replaceChildren();
  for (const r of rows) {
    const tr = el("tr", "hover cursor-pointer");
    tr.tabIndex = 0;
    tr.append(el("td", "font-mono", r.resi), el("td", "", r.tujuan), el("td", "", r.status));
    tr.addEventListener("click", () => lacak(r.resi));
    tr.addEventListener("keydown", (e) => { if (e.key === "Enter") lacak(r.resi); });
    daftar.append(tr);
  }
}

muatDaftar();
</script></body></html>`

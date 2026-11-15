package main

import (
	"errors"
	"sort"
	"sync"
	"time"
)

var (
	ErrSudahAda    = errors.New("resi sudah ada")
	ErrTidakKetemu = errors.New("resi tidak ditemukan")
)

type Langkah struct {
	Status string    `json:"status"`
	Lokasi string    `json:"lokasi"`
	Waktu  time.Time `json:"waktu"`
}

type Paket struct {
	Resi    string    `json:"resi"`
	Tujuan  string    `json:"tujuan"`
	Riwayat []Langkah `json:"riwayat"`
}

type Ringkas struct {
	Resi   string    `json:"resi"`
	Tujuan string    `json:"tujuan"`
	Status string    `json:"status"`
	Waktu  time.Time `json:"waktu"`
}

type Store struct {
	mu    sync.RWMutex
	paket map[string]*Paket
}

func NewStore() *Store {
	return &Store{paket: map[string]*Paket{}}
}

func (s *Store) Buat(resi, tujuan string) (Paket, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ada := s.paket[resi]; ada {
		return Paket{}, ErrSudahAda
	}
	p := &Paket{Resi: resi, Tujuan: tujuan, Riwayat: []Langkah{}}
	s.paket[resi] = p
	return salin(p), nil
}

func (s *Store) Catat(resi string, langkah Langkah) (Paket, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ada := s.paket[resi]
	if !ada {
		return Paket{}, ErrTidakKetemu
	}
	langkah.Waktu = time.Now().UTC()
	p.Riwayat = append(p.Riwayat, langkah)
	return salin(p), nil
}

func (s *Store) Lacak(resi string) (Paket, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ada := s.paket[resi]
	if !ada {
		return Paket{}, ErrTidakKetemu
	}
	return salin(p), nil
}

func (s *Store) Daftar() []Ringkas {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Ringkas, 0, len(s.paket))
	for _, p := range s.paket {
		r := Ringkas{Resi: p.Resi, Tujuan: p.Tujuan, Status: "Belum dipindai"}
		if n := len(p.Riwayat); n > 0 {
			r.Status = p.Riwayat[n-1].Status
			r.Waktu = p.Riwayat[n-1].Waktu
		}
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Resi < out[j].Resi })
	return out
}

func (s *Store) Muat(p Paket) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := salin(&p)
	s.paket[cp.Resi] = &cp
}

func salin(p *Paket) Paket {
	out := *p
	out.Riwayat = append([]Langkah{}, p.Riwayat...)
	return out
}

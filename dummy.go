package main

import "time"

func SeedDummy(s *Store) {
	for _, p := range dummyPaket(time.Now().UTC()) {
		s.Muat(p)
	}
}

func dummyPaket(now time.Time) []Paket {
	jam := func(n int) time.Time { return now.Add(-time.Duration(n) * time.Hour) }
	return []Paket{
		{Resi: "SHP1001", Tujuan: "Padang", Riwayat: []Langkah{
			{Status: "Pesanan dibuat", Lokasi: "Jakarta Selatan", Waktu: jam(72)},
			{Status: "Dijemput kurir", Lokasi: "Jakarta Selatan", Waktu: jam(60)},
			{Status: "Tiba di hub", Lokasi: "Hub Jakarta", Waktu: jam(48)},
			{Status: "Dalam perjalanan", Lokasi: "Hub Padang", Waktu: jam(20)},
			{Status: "Diantar kurir", Lokasi: "Padang", Waktu: jam(2)},
		}},
		{Resi: "SHP1002", Tujuan: "Pekanbaru", Riwayat: []Langkah{
			{Status: "Pesanan dibuat", Lokasi: "Bandung", Waktu: jam(30)},
			{Status: "Dijemput kurir", Lokasi: "Bandung", Waktu: jam(24)},
		}},
		{Resi: "SHP1003", Tujuan: "Solok", Riwayat: []Langkah{
			{Status: "Pesanan dibuat", Lokasi: "Surabaya", Waktu: jam(120)},
			{Status: "Dalam perjalanan", Lokasi: "Hub Surabaya", Waktu: jam(96)},
			{Status: "Tiba di hub", Lokasi: "Hub Padang", Waktu: jam(40)},
			{Status: "Diterima penerima", Lokasi: "Solok", Waktu: jam(6)},
		}},
	}
}

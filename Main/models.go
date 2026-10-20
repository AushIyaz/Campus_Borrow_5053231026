package main

import (
	"sync"
)

type SumberDaya struct {
	ID    int    `json:"id"`
	Nama  string `json:"nama"`
	Jenis string `json:"jenis"`
	Stok  int    `json:"stok"`
}

type Peminjaman struct {
	ID            int    `json:"id"`
	SumberDayaID  int    `json:"sumber_daya_id"`
	Peminjam      string `json:"peminjam"`
	Tanggal       string `json:"tanggal"`
	SlotMulai     int    `json:"slot_mulai"`
	SlotSelesai   int    `json:"slot_selesai"`
	Jumlah        int    `json:"jumlah"`
	Status        string `json:"status"`
}

type AppMemory struct {
	mu           sync.Mutex
	sumberDayas  []SumberDaya
	peminjamans  []Peminjaman
	nextSDID     int
	nextPeminjamID int
}

var db = &AppMemory{
	sumberDayas:    make([]SumberDaya, 0),
	peminjamans:    make([]Peminjaman, 0),
	nextSDID:       1,
	nextPeminjamID: 1,
}
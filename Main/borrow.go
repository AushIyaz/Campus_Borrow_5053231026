package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type PeminjamanInput struct {
	SumberDayaID int    `json:"sumber_daya_id"`
	Peminjam     string `json:"peminjam"`
	Tanggal      string `json:"tanggal"`
	SlotMulai    int    `json:"slot_mulai"`
	SlotSelesai  int    `json:"slot_selesai"`
	Jumlah       int    `json:"jumlah"`
}

func handlePostPeminjaman(w http.ResponseWriter, r *http.Request) {
	var input PeminjamanInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Body tidak valid"})
		return
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	p := Peminjaman{
		ID:           db.nextPeminjamID,
		SumberDayaID: input.SumberDayaID,
		Peminjam:     strings.TrimSpace(input.Peminjam),
		Tanggal:      input.Tanggal,
		SlotMulai:    input.SlotMulai,
		SlotSelesai:  input.SlotSelesai,
		Jumlah:       input.Jumlah,
		Status:       "aktif",
	}
	db.nextPeminjamID++
	db.peminjamans = append(db.peminjamans, p)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

func handleGetPeminjaman(w http.ResponseWriter, r *http.Request) {
	db.mu.Lock()
	defer db.mu.Unlock()

	result := db.peminjamans
	if result == nil {
		result = []Peminjaman{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func handleGetPeminjamanDetail(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ID tidak valid"})
		return
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	for _, p := range db.peminjamans {
		if p.ID == id {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(p)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{"error": "Peminjaman tidak ditemukan"})
}
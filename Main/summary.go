package main

import (
	"encoding/json"
	"net/http"
	"time"
)

type SumberDayaRingkasan struct {
	SumberDayaID     int    `json:"sumber_daya_id"`
	Nama             string `json:"nama"`
	JumlahPeminjaman int    `json:"jumlah_peminjaman"`
}

type RingkasanResponse struct {
	Tanggal          string               `json:"tanggal"`
	JumlahPeminjaman int                  `json:"jumlah_peminjaman"`
	TotalUnitSlot    int                  `json:"total_unit_slot"`
	PerSumberDaya    []SumberDayaRingkasan `json:"per_sumber_daya"`
}

func handleGetRingkasan(w http.ResponseWriter, r *http.Request) {
	tanggal := r.URL.Query().Get("tanggal")
	if tanggal == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Parameter tanggal wajib diisi"})
		return
	}

	if _, err := time.Parse("2006-01-02", tanggal); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Format tanggal tidak valid (YYYY-MM-DD)"})
		return
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	jumlahPeminjamanTotal := 0
	totalUnitSlot := 0

	peminjamanPerSD := make(map[int]int)

	for _, p := range db.peminjamans {
		if p.Status == "aktif" && p.Tanggal == tanggal {
			jumlahPeminjamanTotal++
			
			banyakSlot := p.SlotSelesai - p.SlotMulai + 1
			totalUnitSlot += p.Jumlah * banyakSlot

			peminjamanPerSD[p.SumberDayaID]++
		}
	}

	var perSumberDaya []SumberDayaRingkasan
	for _, sd := range db.sumberDayas {
		jml := peminjamanPerSD[sd.ID]
		perSumberDaya = append(perSumberDaya, SumberDayaRingkasan{
			SumberDayaID:     sd.ID,
			Nama:             sd.Nama,
			JumlahPeminjaman: jml,
		})
	}

	if perSumberDaya == nil {
		perSumberDaya = []SumberDayaRingkasan{}
	}

	resp := RingkasanResponse{
		Tanggal:          tanggal,
		JumlahPeminjaman: jumlahPeminjamanTotal,
		TotalUnitSlot:    totalUnitSlot,
		PerSumberDaya:    perSumberDaya,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
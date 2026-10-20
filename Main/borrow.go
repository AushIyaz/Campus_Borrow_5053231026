package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type PeminjamanInput struct {
	SumberDayaID int    `json:"sumber_daya_id"`
	Peminjam     string `json:"peminjam"`
	Tanggal      string `json:"tanggal"`
	SlotMulai    int    `json:"slot_mulai"`
	SlotSelesai  int    `json:"slot_selesai"`
	Jumlah       int    `json:"jumlah"`
}

func validateInput(input PeminjamanInput) (string, int) {
	peminjam := strings.TrimSpace(input.Peminjam)
	if peminjam == "" || len([]rune(peminjam)) > 50 {
		return "Peminjam wajib, maksimal 50 karakter", http.StatusBadRequest
	}

	if _, err := time.Parse("2006-01-02", input.Tanggal); err != nil {
		return "Format tanggal tidak valid (YYYY-MM-DD)", http.StatusBadRequest
	}

	if input.SlotMulai < 1 || input.SlotMulai > 8 || input.SlotSelesai < 1 || input.SlotSelesai > 8 || input.SlotMulai > input.SlotSelesai {
		return "Rentang slot harus antara 1 sampai 8 dan mulai <= selesai", http.StatusBadRequest
	}

	if (input.SlotSelesai - input.SlotMulai + 1) > 3 {
		return "Durasi peminjaman maksimal adalah 3 slot", http.StatusBadRequest
	}

	if input.Jumlah < 1 {
		return "Jumlah minimal 1", http.StatusBadRequest
	}

	sdFound := false
	var stokMax int
	for _, sd := range db.sumberDayas {
		if sd.ID == input.SumberDayaID {
			sdFound = true
			stokMax = sd.Stok
			break
		}
	}
	if !sdFound {
		return "Sumber daya tidak ditemukan", http.StatusBadRequest
	}

	if input.Jumlah > stokMax {
		return "Jumlah melebihi stok sumber daya", http.StatusBadRequest
	}

	return "", 0
}

func countActiveKuota(peminjam string, tanggal string, excludeID int) int {
	count := 0
	pTrim := strings.TrimSpace(peminjam)
	for _, p := range db.peminjamans {
		if p.ID != excludeID && p.Status == "aktif" && strings.TrimSpace(p.Peminjam) == pTrim && p.Tanggal == tanggal {
			count++
		}
	}
	return count
}

func checkBentrok(sumberDayaID int, tanggal string, slotMulai int, slotSelesai int, jumlahBaru int, excludeID int) bool {
	for _, p := range db.peminjamans {
		if p.ID != excludeID && p.Status == "aktif" && p.SumberDayaID == sumberDayaID && p.Tanggal == tanggal {
			if !(p.SlotSelesai < slotMulai || p.SlotMulai > slotSelesai) {
			}
		}
	}

	for s := slotMulai; s <= slotSelesai; s++ {
		terpakai := 0
		for _, p := range db.peminjamans {
			if p.ID != excludeID && p.Status == "aktif" && p.SumberDayaID == sumberDayaID && p.Tanggal == tanggal {
				if s >= p.SlotMulai && s <= p.SlotSelesai {
					terpakai += p.Jumlah
				}
			}
		}

		stok := 0
		for _, sd := range db.sumberDayas {
			if sd.ID == sumberDayaID {
				stok = sd.Stok
				break
			}
		}

		if terpakai+jumlahBaru > stok {
			return true
		}
	}
	return false
}

func handlePostPeminjaman(w http.ResponseWriter, r *http.Request) {
	var input PeminjamanInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Body tidak valid"})
		return
	}

	if errMsg, code := validateInput(input); code != 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(map[string]string{"error": errMsg})
		return
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	peminjamTrimmed := strings.TrimSpace(input.Peminjam)

	currentKuota := countActiveKuota(peminjamTrimmed, input.Tanggal, 0)
	if currentKuota >= 2 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": "Kuota peminjaman harian telah penuh (maksimal 2)"})
		return
	}

	if checkBentrok(input.SumberDayaID, input.Tanggal, input.SlotMulai, input.SlotSelesai, input.Jumlah, 0) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": "Slot bentrok atau stok tidak mencukupi"})
		return
	}

	p := Peminjaman{
		ID:           db.nextPeminjamID,
		SumberDayaID: input.SumberDayaID,
		Peminjam:     peminjamTrimmed,
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
	query := r.URL.Query()
	filterSDStr := query.Get("sumber_daya_id")
	filterTanggal := query.Get("tanggal")
	filterStatus := query.Get("status")

	var filterSD int = -1
	var err error
	if filterSDStr != "" {
		filterSD, err = strconv.Atoi(filterSDStr)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "sumber_daya_id bukan angka"})
			return
		}
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	var result []Peminjaman = []Peminjaman{}
	for _, p := range db.peminjamans {
		match := true
		if filterSD != -1 && p.SumberDayaID != filterSD {
			match = false
		}
		if filterTanggal != "" && p.Tanggal != filterTanggal {
			match = false
		}
		if filterStatus != "" && p.Status != filterStatus {
			match = false
		}
		if match {
			result = append(result, p)
		}
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

func handlePutPeminjaman(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ID bukan angka"})
		return
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	var oldIndex = -1
	for i, p := range db.peminjamans {
		if p.ID == id {
			oldIndex = i
			break
		}
	}
	if oldIndex == -1 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Peminjaman tidak ditemukan"})
		return
	}

	var input PeminjamanInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Body tidak valid"})
		return
	}

	if errMsg, code := validateInput(input); code != 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(map[string]string{"error": errMsg})
		return
	}

	peminjamTrimmed := strings.TrimSpace(input.Peminjam)

	currentKuota := countActiveKuota(peminjamTrimmed, input.Tanggal, id)
	if currentKuota >= 2 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": "Kuota peminjaman harian peminjam sudah penuh"})
		return
	}

	if checkBentrok(input.SumberDayaID, input.Tanggal, input.SlotMulai, input.SlotSelesai, input.Jumlah, id) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": "Slot bentrok atau stok tidak mencukupi"})
		return
	}

	db.peminjamans[oldIndex].SumberDayaID = input.SumberDayaID
	db.peminjamans[oldIndex].Peminjam = peminjamTrimmed
	db.peminjamans[oldIndex].Tanggal = input.Tanggal
	db.peminjamans[oldIndex].SlotMulai = input.SlotMulai
	db.peminjamans[oldIndex].SlotSelesai = input.SlotSelesai
	db.peminjamans[oldIndex].Jumlah = input.Jumlah

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(db.peminjamans[oldIndex])
}

func handleDelPeminjaman(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ID bukan angka"})
		return
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	foundIndex := -1
	for i, p := range db.peminjamans {
		if p.ID == id {
			foundIndex = i
			break
		}
	}

	if foundIndex == -1 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Peminjaman tidak ditemukan"})
		return
	}

	db.peminjamans = append(db.peminjamans[:foundIndex], db.peminjamans[foundIndex+1:]...)

	w.WriteHeader(http.StatusNoContent)
}

func handleGetKuotaPeminjam(w http.ResponseWriter, r *http.Request) {
	nama := r.PathValue("nama")
	tanggal := r.URL.Query().Get("tanggal")

	if tanggal == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Parameter tanggal wajib diisi dan valid"})
		return
	}

	if _, err := time.Parse("2006-01-02", tanggal); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Format tanggal tidak valid"})
		return
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	terpakai := countActiveKuota(nama, tanggal, 0)
	maksimal := 2
	sisa := maksimal - terpakai
	if sisa < 0 {
		sisa = 0
	}

	response := map[string]interface{}{
		"peminjam":  strings.TrimSpace(nama),
		"tanggal":   tanggal,
		"terpakai":  terpakai,
		"maksimal":  maksimal,
		"sisa":      sisa,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
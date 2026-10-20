package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func handlePostSumberDaya(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Nama  string `json:"nama"`
		Jenis string `json:"jenis"`
		Stok  *int   `json:"stok"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Body tidak valid"})
		return
	}

	namaTrimmed := strings.TrimSpace(input.Nama)
	if namaTrimmed == "" || len([]rune(namaTrimmed)) > 50 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Nama wajib, maksimal 50 karakter"})
		return
	}

	if input.Jenis != "ruang" && input.Jenis != "alat" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Jenis harus 'ruang' atau 'alat'"})
		return
	}

	if input.Stok == nil || *input.Stok < 1 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Stok minimal 1"})
		return
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	sd := SumberDaya{
		ID:    db.nextSDID,
		Nama:  namaTrimmed,
		Jenis: input.Jenis,
		Stok:  *input.Stok,
	}
	db.nextSDID++
	db.sumberDayas = append(db.sumberDayas, sd)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sd)
}

func handleGetSumberDaya(w http.ResponseWriter, r *http.Request) {
	db.mu.Lock()
	defer db.mu.Unlock()

	result := db.sumberDayas
	if result == nil {
		result = []SumberDaya{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func handleGetSumberDayaDetail(w http.ResponseWriter, r *http.Request) {
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

	for _, sd := range db.sumberDayas {
		if sd.ID == id {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(sd)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{"error": "Sumber daya tidak ditemukan"})
}
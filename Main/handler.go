package main

import (
	"net/http"
)

func NewHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /sumber-daya", handlePostSumberDaya)
	mux.HandleFunc("GET /sumber-daya", handleGetSumberDaya)
	mux.HandleFunc("GET /sumber-daya/{id}", handleGetSumberDayaDetail)

	mux.HandleFunc("POST /peminjaman", handlePostPeminjaman)
	mux.HandleFunc("GET /peminjaman", handleGetPeminjaman)
	mux.HandleFunc("GET /peminjaman/{id}", handleGetPeminjamanDetail)
	mux.HandleFunc("PUT /peminjaman/{id}", handlePutPeminjaman)
	mux.HandleFunc("DELETE /peminjaman/{id}", handleDelPeminjaman)

	mux.HandleFunc("GET /peminjam/{nama}/kuota", handleGetKuotaPeminjam)

	return mux
}
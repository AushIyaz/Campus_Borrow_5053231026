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

	return mux
}
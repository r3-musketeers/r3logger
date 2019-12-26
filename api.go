package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func createRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/join", joinHandler).
		Methods("POST").
		Schemes("", "http")

	r.HandleFunc("/recover", recoverHandler).
		Methods("GET").
		Schemes("", "http")

	r.HandleFunc("/truncate", truncateHandler).
		Methods("PUT").
		Schemes("", "http")
	return r
}

// TODO: modify current error handling on API endpoints. Must create a
// user-friendly error structure with generic error events, to abstract
// internal service implementation from clients, and just record the full
// error messages into a server log for debuging.
func joinHandler(w http.ResponseWriter, r *http.Request) {
}

func recoverHandler(w http.ResponseWriter, r *http.Request) {
}

func truncateHandler(w http.ResponseWriter, r *http.Request) {
}

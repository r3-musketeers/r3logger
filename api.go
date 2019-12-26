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

// joinHandler initializes the handshake for connection establishment,
// sending a JOIN request to the informed Raft leader address.
func joinHandler(w http.ResponseWriter, r *http.Request) {
}

// recovHandler returns a state file, bounded by the requested log interval.
func recoverHandler(w http.ResponseWriter, r *http.Request) {
}

// truncateHandler tries to truncate a logFile, eliminating all commands older
// than the desired index. Truncate is only commited if, and only if, a majority
// of non-faulty replicas requests a common log interval.
func truncateHandler(w http.ResponseWriter, r *http.Request) {
}

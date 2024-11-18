package main

import (
	"filepack/internal/handlers"
	"net/http"
)

func main() {
	// Create a new ServeMux router
	mux := http.NewServeMux()

	// Register routes with the mux router
	mux.HandleFunc("/api/archive/information", handlers.ProcessArchive)
	mux.HandleFunc("/api/archive/files", handlers.CreateArchive)
	mux.HandleFunc("/api/mail/file", handlers.SendFileToEmails)

	// Start the HTTP server with the mux router
	http.ListenAndServe(":8080", mux)
}

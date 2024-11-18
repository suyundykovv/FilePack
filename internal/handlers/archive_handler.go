package handlers

import (
	"encoding/json"
	"filepack/internal/services"
	"net/http"
)

// ProcessArchive handles the extraction of archive details from the uploaded file
func ProcessArchive(w http.ResponseWriter, r *http.Request) {
	// Retrieve the uploaded file and file header
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Extract the archive details using the service
	result, err := services.ExtractArchiveDetails(fileHeader) // Pass the *multipart.FileHeader instead of multipart.File
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Return the extracted details in JSON format
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}

// CreateArchive handles the creation of a ZIP archive from multiple uploaded files
func CreateArchive(w http.ResponseWriter, r *http.Request) {
	// Retrieve the uploaded files
	err := r.ParseMultipartForm(10 << 20) // Limit to 10 MB
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files[]"]
	if len(files) == 0 {
		http.Error(w, "No files provided", http.StatusBadRequest)
		return
	}

	// Create a ZIP archive from the files
	archive, err := services.CreateZipArchive(files)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set the response headers for downloading the ZIP file
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=archive.zip")
	http.ServeFile(w, r, archive.Name())
}

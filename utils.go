package main

import (
	"mime/multipart"
	"net/http"
)

func isValidFileType(fileHeader *multipart.FileHeader) bool {
	mimeType := fileHeader.Header.Get("Content-Type")
	validMimeTypes := map[string]bool{
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		"application/xml": true,
		"image/jpeg":      true,
		"image/png":       true,
		"application/pdf": true,
	}

	// Check if MIME type is valid
	return validMimeTypes[mimeType]
}

// Helper function for sending error responses
func sendError(w http.ResponseWriter, errMessage string) {
	http.Error(w, errMessage, http.StatusBadRequest)
}

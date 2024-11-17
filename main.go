package main

import (
	"encoding/json"
	"log"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

// Declare all handlers above main
func handleArchiveInfo(w http.ResponseWriter, r *http.Request) {
	// Parse the uploaded file
	file, _, err := r.FormFile("file")
	if err != nil {
		sendError(w, "Unable to read file")
		return
	}
	defer file.Close()

	// Temporarily save the file
	tmpFile, err := os.CreateTemp("", "upload-*.zip")
	if err != nil {
		sendError(w, "Unable to create temp file")
		return
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.ReadFrom(file)
	if err != nil {
		sendError(w, "Unable to save uploaded file")
		return
	}

	// Extract archive contents
	files, err := extractArchive(tmpFile.Name())
	if err != nil {
		sendError(w, err.Error())
		return
	}

	// Prepare response
	var totalSize int64
	var totalFiles int
	for _, f := range files {
		totalSize += f.Size
		totalFiles++
	}

	response := ArchiveInfo{
		Filename:    tmpFile.Name(),
		ArchiveSize: totalSize,
		TotalSize:   totalSize,
		TotalFiles:  totalFiles,
		Files:       files,
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleCreateArchive(w http.ResponseWriter, r *http.Request) {
	// Parse files from the request
	err := r.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		sendError(w, "Unable to parse files")
		return
	}

	var files []*multipart.FileHeader
	for _, headers := range r.MultipartForm.File {
		for _, fileHeader := range headers {
			if !isValidFileType(fileHeader) {
				sendError(w, "Invalid file type")
				return
			}
			files = append(files, fileHeader)
		}
	}

	// Create a zip archive
	archiveFilePath, err := createZipArchive(files)
	if err != nil {
		sendError(w, "Unable to create archive")
		return
	}

	// Send the archive as the response
	w.Header().Set("Content-Type", "application/zip")
	http.ServeFile(w, r, archiveFilePath)
}

func handleSendEmail(w http.ResponseWriter, r *http.Request) {
	// Parse the email list and file
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		sendError(w, "Unable to parse form")
		return
	}

	emailList := r.Form["email"]
	if len(emailList) == 0 {
		sendError(w, "Email address is required")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		sendError(w, "Unable to read file")
		return
	}

	// Send the email with the attachment
	err = sendEmail(file, emailList)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	// Success response
	w.Write([]byte("Email sent successfully"))
}

func main() {
	r := mux.NewRouter()

	// Routes
	r.HandleFunc("/api/archive/information", handleArchiveInfo).Methods("POST")
	r.HandleFunc("/api/archive/files", handleCreateArchive).Methods("POST")
	r.HandleFunc("/api/mail/file", handleSendEmail).Methods("POST")

	// Start the server
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

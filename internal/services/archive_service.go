package services

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/h2non/filetype" // External library for MIME type detection
)

// ArchiveFile represents a file's details inside an archive
type ArchiveFile struct {
	FilePath string  `json:"file_path"`
	Size     float64 `json:"size"`
	MimeType string  `json:"mimetype"`
}

// ArchiveDetails represents the detailed structure of an archive
type ArchiveDetails struct {
	Filename    string        `json:"filename"`
	ArchiveSize float64       `json:"archive_size"`
	TotalSize   float64       `json:"total_size"`
	TotalFiles  float64       `json:"total_files"`
	Files       []ArchiveFile `json:"files"`
}

// ExtractArchiveDetails processes the uploaded file and extracts its structure details
func ExtractArchiveDetails(file *multipart.FileHeader) (*ArchiveDetails, error) {
	// Open the uploaded file
	f, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer f.Close()

	// Verify if the file is a valid ZIP archive
	buffer := make([]byte, 261)
	if _, err := f.Read(buffer); err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}
	if !filetype.IsArchive(buffer) {
		return nil, errors.New("the uploaded file is not a valid archive")
	}

	// Reset file pointer to the beginning for ZIP reader
	if _, err := f.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to reset file pointer: %v", err)
	}

	// Open the ZIP archive
	zipReader, err := zip.NewReader(f, file.Size)
	if err != nil {
		return nil, fmt.Errorf("failed to read archive: %v", err)
	}

	// Process files in the archive
	var files []ArchiveFile
	var totalSize float64
	for _, zipFile := range zipReader.File {
		totalSize += float64(zipFile.UncompressedSize64)

		// Open each file to detect MIME type
		zipFileReader, err := zipFile.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to read file in archive: %v", err)
		}

		mimeBuffer := make([]byte, 512)
		_, _ = zipFileReader.Read(mimeBuffer)
		zipFileReader.Close()

		mimeType := http.DetectContentType(mimeBuffer)

		files = append(files, ArchiveFile{
			FilePath: zipFile.Name,
			Size:     float64(zipFile.UncompressedSize64),
			MimeType: mimeType,
		})
	}

	// Construct and return archive details
	return &ArchiveDetails{
		Filename:    file.Filename,
		ArchiveSize: float64(file.Size),
		TotalSize:   totalSize,
		TotalFiles:  float64(len(files)),
		Files:       files,
	}, nil
}

// CreateZipArchive creates a ZIP archive from the given files
func CreateZipArchive(files []*multipart.FileHeader) (*os.File, error) {
	// Create a buffer to hold the ZIP archive
	var buffer bytes.Buffer
	zipWriter := zip.NewWriter(&buffer)

	// Allowed MIME types for the files
	allowedMIMETypes := map[string]bool{
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		"application/xml": true,
		"image/jpeg":      true,
		"image/png":       true,
	}

	// Process each file
	for _, file := range files {
		// Open the file
		f, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open file %s: %v", file.Filename, err)
		}

		// Read MIME type
		mimeBuffer := make([]byte, 512)
		_, _ = f.Read(mimeBuffer)
		f.Seek(0, 0) // Reset file pointer

		mimeType := http.DetectContentType(mimeBuffer)

		// Validate MIME type
		if !allowedMIMETypes[mimeType] {
			return nil, fmt.Errorf("invalid file type: %s (file: %s)", mimeType, file.Filename)
		}

		// Add file to the ZIP archive
		zipFileWriter, err := zipWriter.Create(file.Filename)
		if err != nil {
			return nil, fmt.Errorf("failed to add file %s to archive: %v", file.Filename, err)
		}

		// Copy file content to zip file writer
		_, err = io.Copy(zipFileWriter, f)
		if err != nil {
			return nil, fmt.Errorf("failed to write file %s to archive: %v", file.Filename, err)
		}
		f.Close()
	}

	// Close the ZIP writer
	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to finalize archive: %v", err)
	}

	// Save the ZIP archive to a temporary file
	tempFile, err := os.CreateTemp("", "archive-*.zip")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %v", err)
	}

	if _, err := tempFile.Write(buffer.Bytes()); err != nil {
		tempFile.Close()
		os.Remove(tempFile.Name())
		return nil, fmt.Errorf("failed to write archive to temporary file: %v", err)
	}

	tempFile.Seek(0, 0) // Reset file pointer
	return tempFile, nil
}

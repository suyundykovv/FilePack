package main

import (
	"archive/zip"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

type File struct {
	FilePath string `json:"file_path"`
	Size     int64  `json:"size"`
	MimeType string `json:"mimetype"`
}

type ArchiveInfo struct {
	Filename    string `json:"filename"`
	ArchiveSize int64  `json:"archive_size"`
	TotalSize   int64  `json:"total_size"`
	TotalFiles  int    `json:"total_files"`
	Files       []File `json:"files"`
}

// Extract archive contents
func extractArchive(archivePath string) ([]File, error) {
	var files []File

	// Open the archive
	archiveFile, err := os.Open(archivePath)
	if err != nil {
		return nil, fmt.Errorf("could not open archive: %v", err)
	}
	defer archiveFile.Close()

	// Open the ZIP reader
	zipReader, err := zip.OpenReader(archiveFile.Name())
	if err != nil {
		return nil, fmt.Errorf("could not open ZIP archive: %v", err)
	}
	defer zipReader.Close()

	// Iterate over each file in the archive
	for _, file := range zipReader.File {
		files = append(files, File{
			FilePath: file.Name,
			Size:     file.FileInfo().Size(),
			MimeType: "application/zip", // assuming ZIP as MIME, modify based on file type
		})
	}
	return files, nil
}

// Create ZIP archive
func createZipArchive(files []*multipart.FileHeader) (string, error) {
	// Create temporary file for the archive
	archiveFileName := fmt.Sprintf("archive-%d.zip", time.Now().Unix())
	archiveFilePath := filepath.Join("temp", archiveFileName)
	zipFile, err := os.Create(archiveFilePath)
	if err != nil {
		return "", fmt.Errorf("unable to create zip file: %v", err)
	}
	defer zipFile.Close()

	// Create a new ZIP writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Add files to the ZIP archive
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			return "", fmt.Errorf("unable to open file: %v", err)
		}
		defer file.Close()

		// Create a new file in the archive
		zipFileWriter, err := zipWriter.Create(fileHeader.Filename)
		if err != nil {
			return "", fmt.Errorf("unable to create file in archive: %v", err)
		}

		// Copy file contents to the archive
		_, err = io.Copy(zipFileWriter, file)
		if err != nil {
			return "", fmt.Errorf("error while copying file to archive: %v", err)
		}
	}

	return archiveFilePath, nil
}

package handlers

import (
	"bytes"
	"filepack/internal/services"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func SendFileToEmails(w http.ResponseWriter, r *http.Request) {
	// Retrieve the uploaded file
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File is required", http.StatusBadRequest)
		return
	}
	defer file.Close() // Ensure the file is closed after use

	// Retrieve the list of recipient emails
	emails := r.FormValue("emails")
	if emails == "" {
		http.Error(w, "Emails are required", http.StatusBadRequest)
		return
	}

	// Convert the comma-separated emails into a slice of strings
	emailList := strings.Split(emails, ",")

	// Read the file content into a byte slice (e.g., to send via email)
	fileContent, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	// Prepare the email request
	emailRequest := services.EmailRequest{
		SenderEmail:     "sender@example.com",         // Update with the sender email
		RecipientEmails: emailList,                    // List of recipient emails
		Subject:         "Subject here",               // Email subject
		Body:            "Email body here",            // Email body
		Attachment:      bytes.NewReader(fileContent), // File content as an attachment
		Filename:        "attached_file.txt",          // Attachment filename
	}

	// Send the email
	err = services.SendEmail(emailRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Return a success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "File sent successfully"}`)
}

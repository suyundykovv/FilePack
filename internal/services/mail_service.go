package services

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/smtp"
	"net/textproto"
	"os"
	"strings"
)

type EmailRequest struct {
	SenderEmail     string
	RecipientEmails []string
	Subject         string
	Body            string
	Attachment      io.Reader // Use io.Reader for the file content
	Filename        string    // File name for the attachment
}

func ValidateEmail(email string) error {
	// Add email validation logic
	return nil
}

func ValidateEmailList(emails []string) error {
	// Validate email list
	for _, email := range emails {
		if err := ValidateEmail(email); err != nil {
			return err
		}
	}
	return nil
}

func stringJoinEmails(emails []string) string {
	return strings.Join(emails, ", ")
}

// SendEmail sends an email with optional attachment
func SendEmail(request EmailRequest) error {
	// Validate sender email
	if err := ValidateEmail(request.SenderEmail); err != nil {
		return err
	}

	// Validate recipient emails
	if err := ValidateEmailList(request.RecipientEmails); err != nil {
		return err
	}

	// SMTP server configuration
	smtpHost := "smtp.gmail.com" // Example: Gmail's SMTP server
	smtpPort := "587"            // Example: TLS port

	// Environment variables for credentials
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	if smtpUsername == "" || smtpPassword == "" {
		return errors.New("SMTP credentials not set in environment variables")
	}

	// Create a buffer to write the email content
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	// Create the email headers
	headers := textproto.MIMEHeader{} // Use MIMEHeader from textproto
	headers.Set("From", request.SenderEmail)
	headers.Set("To", stringJoinEmails(request.RecipientEmails))
	headers.Set("Subject", request.Subject)

	// Write the email headers
	for key, value := range headers {
		buffer.WriteString(fmt.Sprintf("%s: %s\r\n", key, value[0]))
	}

	// Write the message body
	bodyPart, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Type":              []string{"text/plain; charset=UTF-8"},
		"Content-Transfer-Encoding": []string{"base64"},
	})
	if err != nil {
		return fmt.Errorf("failed to create body part: %v", err)
	}
	_, err = bodyPart.Write([]byte(request.Body))
	if err != nil {
		return fmt.Errorf("failed to write body: %v", err)
	}

	// Add attachment if available
	if request.Attachment != nil && request.Filename != "" {
		attachmentPart, err := writer.CreatePart(textproto.MIMEHeader{
			"Content-Type":              []string{"application/octet-stream"},
			"Content-Disposition":       []string{fmt.Sprintf("attachment; filename=\"%s\"", request.Filename)},
			"Content-Transfer-Encoding": []string{"base64"},
		})
		if err != nil {
			return fmt.Errorf("failed to create attachment part: %v", err)
		}
		_, err = io.Copy(attachmentPart, request.Attachment)
		if err != nil {
			return fmt.Errorf("failed to copy attachment: %v", err)
		}
	}

	// Close the multipart writer to finalize the message
	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close multipart writer: %v", err)
	}

	// Connect to the SMTP server and send the email
	auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)
	err = smtp.SendMail(
		smtpHost+":"+smtpPort,
		auth,
		request.SenderEmail,
		request.RecipientEmails,
		buffer.Bytes(),
	)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

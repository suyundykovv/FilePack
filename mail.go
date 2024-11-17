package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/smtp"
)

func sendEmail(file multipart.File, emailList []string) error {
	// Set up the SMTP client (use your own credentials here or environment variables)
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	auth := smtp.PlainAuth("", "your-email@gmail.com", "your-email-password", smtpHost)

	// Prepare email headers
	subject := "File attachment"
	from := "your-email@gmail.com"
	to := emailList
	headers := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n", from, emailList[0], subject)

	// Read the file content
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("unable to read file content: %v", err)
	}

	// Prepare email body
	body := bytes.NewBuffer(nil)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "attachment")
	if err != nil {
		return fmt.Errorf("unable to create form file: %v", err)
	}
	_, err = part.Write(fileBytes)
	if err != nil {
		return fmt.Errorf("error writing to part: %v", err)
	}
	writer.Close()

	// Compose email and send
	msg := append([]byte(headers), body.Bytes()...)
	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, msg)
	if err != nil {
		return fmt.Errorf("unable to send email: %v", err)
	}

	log.Println("Email sent successfully")
	return nil
}

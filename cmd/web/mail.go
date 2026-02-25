package main

import (
	"bytes"
	"html/template"
	"net/smtp"
	"os"
)

type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

// sendHTMLEmail sends an email with a pre-built HTML body string.
func sendHTMLEmail(to, subject, htmlBody string) error {
	cfg := SMTPConfig{
		Port:     os.Getenv("MAIL_PORT"),
		From:     os.Getenv("MAIL_FROM"),
		Host:     os.Getenv("MAIL_HOST"),
		Username: os.Getenv("MAIL_USERNAME"),
		Password: os.Getenv("MAIL_PASSWORD"),
	}
	msg := "MIME-Version: 1.0\r\n"
	msg += "Content-Type: text/html; charset=\"UTF-8\"\r\n"
	msg += "From: FaithConnect <" + cfg.From + ">\r\n"
	msg += "To: " + to + "\r\n"
	msg += "Subject: " + subject + "\r\n\r\n"
	msg += htmlBody
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	return smtp.SendMail(cfg.Host+":"+cfg.Port, auth, cfg.From, []string{to}, []byte(msg))
}

func SendEmail(config SMTPConfig, templateFile string, data any, to string, subject string) error {
	config.Port = os.Getenv("MAIL_PORT")
	config.From = os.Getenv("MAIL_FROM")
	config.Host = os.Getenv("MAIL_HOST")
	config.Username = os.Getenv("MAIL_USERNAME")
	config.Password = os.Getenv("MAIL_PASSWORD")

	// 1. Parse Template
	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		return err
	}

	// 2. Inject data into template
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return err
	}

	// 3. Prepare email headers and body
	msg := "MIME-Version: 1.0\r\n"
	msg += "Content-Type: text/html; charset=\"UTF-8\"\r\n"
	msg += "From: " + config.From + "\r\n"
	msg += "To: " + to + "\r\n"
	msg += "Subject: " + subject + "\r\n\r\n"
	msg += body.String()

	// 4. Authentication
	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)

	// 5. Send Email
	addr := config.Host + ":" + config.Port
	return smtp.SendMail(addr, auth, config.From, []string{to}, []byte(msg))
}

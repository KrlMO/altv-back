package mail

import (
	"fmt"
	"net/smtp"
)

type SMTPMailer struct {
	host     string
	port     string
	username string
	password string
	from     string
	baseURL  string
}

func NewSMTPMailer(host, port, username, password, from, baseURL string) *SMTPMailer {
	return &SMTPMailer{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
		baseURL:  baseURL,
	}
}

func (m *SMTPMailer) SendVerificationEmail(to string, token string) error {
	auth := smtp.PlainAuth("", m.username, m.password, m.host)

	link := fmt.Sprintf("%s/verify-email?token=%s", m.baseURL, token)

	message := []byte(
		"Subject: Verify your email\r\n" +
			"\r\n" +
			"Click the link to verify your account:\r\n" +
			link,
	)

	return smtp.SendMail(
		m.host+":"+m.port,
		auth,
		m.from,
		[]string{to},
		message,
	)
}

package utils

import (
	"gopkg.in/gomail.v2"
)

type Mailer interface {
	Send(to, subject, body string) error
}

type SMTPMailer struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

func (m *SMTPMailer) Send(to, subject, body string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", m.From)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body)

	dial := gomail.NewDialer(m.Host, m.Port, m.Username, m.Password)
	return dial.DialAndSend(msg)
}

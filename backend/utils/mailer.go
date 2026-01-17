package utils

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"os"
)

type Mailer interface {
	Send(to, subject, body string) error
}

type SMTPMailer struct {
	host string
	port string
	user string
	pass string
	from string
}

func NewSMTPMailerFromEnv() (*SMTPMailer, error) {
	m := &SMTPMailer{
		host: os.Getenv("SMTP_HOST"),
		port: os.Getenv("SMTP_PORT"),
		user: os.Getenv("SMTP_USER"),
		pass: os.Getenv("SMTP_PASS"),
		from: os.Getenv("MAIL_FROM"),
	}
	if m.host == "" || m.port == "" || m.user == "" || m.pass == "" || m.from == "" {
		return nil, fmt.Errorf("missing smtp env vars (SMTP_HOST/SMTP_PORT/SMTP_USER/SMTP_PASS/MAIL_FROM)")
	}
	return m, nil
}

func (m *SMTPMailer) Send(to, subject, body string) error {
	addr := net.JoinHostPort(m.host, m.port)

	c, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer c.Close()

	// STARTTLS se supportato
	if ok, _ := c.Extension("STARTTLS"); ok {
		tlsCfg := &tls.Config{ServerName: m.host}
		if err := c.StartTLS(tlsCfg); err != nil {
			return err
		}
	}

	auth := smtp.PlainAuth("", m.user, m.pass, m.host)
	if err := c.Auth(auth); err != nil {
		return err
	}

	if err := c.Mail(m.from); err != nil {
		return err
	}
	if err := c.Rcpt(to); err != nil {
		return err
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	msg := ""
	msg += fmt.Sprintf("From: %s\r\n", m.from)
	msg += fmt.Sprintf("To: %s\r\n", to)
	msg += fmt.Sprintf("Subject: %s\r\n", subject)
	msg += "MIME-Version: 1.0\r\n"
	msg += "Content-Type: text/plain; charset=utf-8\r\n"
	msg += "\r\n"
	msg += body

	if _, err := w.Write([]byte(msg)); err != nil {
		_ = w.Close()
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	return c.Quit()
}

package utils

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"strings"
	"time"
)

type Mailer interface {
	Send(to, subject, htmlBody string) error
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

func (m *SMTPMailer) Send(to, subject, htmlBody string) error {
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

	boundary := "BOUNDARY_" + fmt.Sprint(time.Now().UnixNano())

	plainBody := stripHTML(htmlBody)

	msg := ""
	msg += fmt.Sprintf("From: %s\r\n", m.from)
	msg += fmt.Sprintf("To: %s\r\n", to)
	msg += fmt.Sprintf("Subject: %s\r\n", subject)
	msg += "MIME-Version: 1.0\r\n"
	msg += fmt.Sprintf("Content-Type: multipart/alternative; boundary=%s\r\n", boundary)
	msg += "\r\n"

	// Parte text/plain
	msg += fmt.Sprintf("--%s\r\n", boundary)
	msg += "Content-Type: text/plain; charset=utf-8\r\n"
	msg += "\r\n"
	msg += plainBody + "\r\n"

	// Parte text/html
	msg += fmt.Sprintf("--%s\r\n", boundary)
	msg += "Content-Type: text/html; charset=utf-8\r\n"
	msg += "\r\n"
	msg += htmlBody + "\r\n"

	msg += fmt.Sprintf("--%s--\r\n", boundary)

	if _, err := w.Write([]byte(msg)); err != nil {
		_ = w.Close()
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	return c.Quit()
}

// Rimuove HTML in modo semplice per la versione plain text
func stripHTML(s string) string {
	replacer := strings.NewReplacer(
		"<br>", "\n",
		"<br/>", "\n",
		"<br />", "\n",
		"</p>", "\n\n",
	)
	out := replacer.Replace(s)
	out = strings.ReplaceAll(out, "<", "")
	out = strings.ReplaceAll(out, ">", "")
	return strings.TrimSpace(out)
}

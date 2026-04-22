// Package email sends transactional emails via SMTP (Mailhog in dev).
package email

import (
	"fmt"
	"net/smtp"
)

// Sender holds SMTP connection settings.
type Sender struct {
	host string
	port int
	from string
}

// New creates a Sender.
func New(host string, port int, from string) *Sender {
	return &Sender{host: host, port: port, from: from}
}

// SendVerification sends an email-verification link to addr.
func (s *Sender) SendVerification(to, verifyURL string) error {
	subject := "Confirm your Trippier account"
	body := fmt.Sprintf(`Hello,

Click the link below to verify your email address and activate your Trippier account:

%s

The link expires in 24 hours.

— The Trippier team
`, verifyURL)

	return s.send(to, subject, body)
}

func (s *Sender) send(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		s.from, to, subject, body,
	)
	return smtp.SendMail(addr, nil, s.from, []string{to}, []byte(msg))
}

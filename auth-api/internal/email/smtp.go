// Package email sends transactional emails via SMTP (Mailhog in dev).
package email

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"time"
)

// Sender holds SMTP connection settings.
type Sender struct {
	host string
	port int
	from string
	user string
	pass string
}

// New creates a Sender. When user and pass are non-empty, SMTP PlainAuth is used (e.g. Resend on port 587).
func New(host string, port int, from, user, pass string) *Sender {
	return &Sender{host: host, port: port, from: from, user: user, pass: pass}
}

var otpTmpl = template.Must(template.New("otp").Parse(`<!DOCTYPE html>
<html lang="fr">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Code de connexion trippier</title>
</head>
<body style="margin:0;padding:0;background:#0f1a14;font-family:Arial,Helvetica,sans-serif;">

<table width="100%" cellpadding="0" cellspacing="0" style="background:#0f1a14;padding:40px 16px;">
  <tr><td align="center">
    <table width="100%" cellpadding="0" cellspacing="0" style="max-width:580px;background:#111e18;border:1px solid #1e3028;border-radius:12px;overflow:hidden;">

      <!-- Header -->
      <tr>
        <td style="padding:20px 28px;background:#131f1a;border-bottom:1px solid #1e3028;">
          <span style="font-size:15px;font-weight:700;color:#e8f2ed;letter-spacing:-0.01em;">
            trippier<span style="color:#6b8f80;">/</span>api
          </span>
        </td>
      </tr>

      <!-- Body -->
      <tr>
        <td style="padding:40px 28px 32px;">
          <h2 style="margin:0 0 12px;font-size:22px;font-weight:700;color:#e8f2ed;line-height:1.3;">
            Confirmez votre adresse
          </h2>
          <p style="margin:0 0 28px;font-size:14px;line-height:1.6;color:#8aa89e;">
            Utilisez le code ci-dessous pour terminer la connexion avec
            <span style="color:#34d39c;">{{.Email}}</span>.
          </p>

          <!-- OTP box -->
          <table width="100%" cellpadding="0" cellspacing="0">
            <tr>
              <td style="background:#0f1a14;border:1px solid #1e3028;border-left:3px solid #34d39c;border-radius:10px;padding:28px 20px;text-align:center;">
                <div style="font-size:11px;font-family:monospace;text-transform:uppercase;letter-spacing:0.1em;color:#6b8f80;margin-bottom:16px;">Code à usage unique</div>
                <div style="font-size:36px;font-weight:700;font-family:monospace;letter-spacing:12px;color:#e8f2ed;">{{.Code}}</div>
                <div style="margin-top:14px;font-size:12px;font-family:monospace;color:#6b8f80;">&#9679; expire dans 15 minutes</div>
              </td>
            </tr>
          </table>

          <!-- Security note -->
          <table width="100%" cellpadding="0" cellspacing="0" style="margin-top:24px;">
            <tr>
              <td style="background:#131f1a;border:1px solid #1e3028;border-left:3px solid #4a6a5a;border-radius:8px;padding:14px 16px;font-size:13px;line-height:1.5;color:#8aa89e;">
                <strong style="color:#e8f2ed;">Vous n'avez pas demandé ce code ?</strong>
                Ignorez ce message — personne n'accédera à votre compte sans ce code.
              </td>
            </tr>
          </table>
        </td>
      </tr>

      <!-- Footer -->
      <tr>
        <td style="padding:16px 28px;border-top:1px solid #1e3028;text-align:center;font-size:11px;color:#4a6a5a;line-height:1.6;">
          trippier/api est open source —
          <a href="https://github.com/ulysse-mercadal/trippier-public-API" style="color:#6b8f80;">github.com/ulysse-mercadal/trippier-public-API</a>
        </td>
      </tr>

    </table>
  </td></tr>
</table>

</body>
</html>`))

// SendOTPCode sends a 6-digit verification code to addr.
func (s *Sender) SendOTPCode(to, code string) error {
	var buf bytes.Buffer
	if err := otpTmpl.Execute(&buf, struct{ Code, Email string }{Code: code, Email: to}); err != nil {
		return fmt.Errorf("email template: %w", err)
	}
	return s.sendHTML(to, "Votre code de connexion trippier : "+code, buf.String())
}

// sendHTML sends a raw HTML email via SMTP with no authentication (suitable for local relay or Mailhog).
// A 15-second timeout prevents the goroutine from blocking indefinitely on hung SMTP servers.
func (s *Sender) sendHTML(to, subject, html string) error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		s.from, to, subject, html,
	)

	var auth smtp.Auth
	if s.user != "" && s.pass != "" {
		auth = smtp.PlainAuth("", s.user, s.pass, s.host)
	}

	done := make(chan error, 1)
	go func() {
		done <- smtp.SendMail(addr, auth, s.from, []string{to}, []byte(msg))
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(15 * time.Second):
		return fmt.Errorf("smtp: send timeout after 15s")
	}
}

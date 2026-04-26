// Package email sends transactional emails via SMTP (Mailhog in dev).
package email

import (
	"bytes"
	"fmt"
	"html/template"
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

var otpTmpl = template.Must(template.New("otp").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <title>Your Trippier verification code</title>
</head>
<body style="margin:0;padding:0;background:#050505;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',system-ui,sans-serif;">
  <table width="100%" cellpadding="0" cellspacing="0" style="background:#050505;padding:48px 16px;">
    <tr>
      <td align="center">
        <table width="100%" cellpadding="0" cellspacing="0" style="max-width:520px;">

          <!-- Logo -->
          <tr>
            <td style="padding-bottom:32px;text-align:center;">
              <span style="font-size:22px;font-weight:700;color:#e5e5e5;letter-spacing:-0.5px;">
                trip<span style="color:#10b981;">pier</span>
              </span>
            </td>
          </tr>

          <!-- Card -->
          <tr>
            <td style="background:#0f0f0f;border:1px solid #1c1c1c;border-radius:12px;padding:40px 36px;">

              <h1 style="margin:0 0 8px;font-size:20px;font-weight:600;color:#e5e5e5;">
                Your verification code
              </h1>
              <p style="margin:0 0 32px;font-size:14px;color:#6b7280;line-height:1.6;">
                Enter this code on the Trippier sign-up page. It expires in&nbsp;15&nbsp;minutes.
              </p>

              <!-- Code display -->
              <div style="text-align:center;margin-bottom:32px;">
                <span style="display:inline-block;font-size:42px;font-weight:800;letter-spacing:0.25em;color:#10b981;font-family:'SF Mono','Fira Code',monospace;background:#0a1f16;border:1px solid rgba(16,185,129,0.3);border-radius:12px;padding:16px 28px;">
                  {{.Code}}
                </span>
              </div>

              <p style="margin:0;font-size:12px;color:#6b7280;text-align:center;">
                If you didn&#39;t create a Trippier account, you can safely ignore this email.
              </p>
            </td>
          </tr>

          <!-- Footer -->
          <tr>
            <td style="padding-top:24px;text-align:center;font-size:12px;color:#374151;">
              trippier.dev — travel data for builders
            </td>
          </tr>

        </table>
      </td>
    </tr>
  </table>
</body>
</html>`))

// SendOTPCode sends a 6-digit verification code to addr.
func (s *Sender) SendOTPCode(to, code string) error {
	var buf bytes.Buffer
	if err := otpTmpl.Execute(&buf, struct{ Code string }{Code: code}); err != nil {
		return fmt.Errorf("email template: %w", err)
	}
	return s.sendHTML(to, "Your Trippier verification code: "+code, buf.String())
}

// sendHTML sends a raw HTML email via SMTP with no authentication (suitable for local relay or Mailhog).
func (s *Sender) sendHTML(to, subject, html string) error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		s.from, to, subject, html,
	)
	return smtp.SendMail(addr, nil, s.from, []string{to}, []byte(msg))
}

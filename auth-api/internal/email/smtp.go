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

var verifyTmpl = template.Must(template.New("verify").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <title>Confirm your Trippier account</title>
</head>
<body style="margin:0;padding:0;background:#050505;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',system-ui,sans-serif;">
  <table width="100%" cellpadding="0" cellspacing="0" style="background:#050505;padding:48px 16px;">
    <tr>
      <td align="center">
        <table width="100%" cellpadding="0" cellspacing="0" style="max-width:520px;">

          <!-- Logo / wordmark -->
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
                Verify your email address
              </h1>
              <p style="margin:0 0 28px;font-size:14px;color:#6b7280;line-height:1.6;">
                Click the button below to activate your account. The link expires in&nbsp;24&nbsp;hours.
              </p>

              <!-- CTA button -->
              <table cellpadding="0" cellspacing="0" style="margin-bottom:32px;">
                <tr>
                  <td style="background:#10b981;border-radius:8px;">
                    <a href="{{.URL}}"
                       style="display:inline-block;padding:12px 28px;font-size:14px;font-weight:600;color:#000;text-decoration:none;border-radius:8px;">
                      Confirm email address
                    </a>
                  </td>
                </tr>
              </table>

              <!-- Fallback link -->
              <p style="margin:0 0 6px;font-size:12px;color:#6b7280;">
                Or copy this link into your browser:
              </p>
              <p style="margin:0;font-size:11px;color:#6b7280;word-break:break-all;font-family:'SF Mono','Fira Code',monospace;">
                {{.URL}}
              </p>

            </td>
          </tr>

          <!-- Footer -->
          <tr>
            <td style="padding-top:24px;text-align:center;font-size:12px;color:#374151;">
              If you didn&#39;t create a Trippier account, you can safely ignore this email.
            </td>
          </tr>

        </table>
      </td>
    </tr>
  </table>
</body>
</html>`))

// SendVerification sends an email-verification link to addr.
func (s *Sender) SendVerification(to, verifyURL string) error {
	var buf bytes.Buffer
	if err := verifyTmpl.Execute(&buf, struct{ URL string }{URL: verifyURL}); err != nil {
		return fmt.Errorf("email template: %w", err)
	}
	return s.sendHTML(to, "Confirm your Trippier account", buf.String())
}

func (s *Sender) sendHTML(to, subject, html string) error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		s.from, to, subject, html,
	)
	return smtp.SendMail(addr, nil, s.from, []string{to}, []byte(msg))
}

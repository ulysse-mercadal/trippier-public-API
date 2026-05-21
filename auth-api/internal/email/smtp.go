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

// Colors — oklch → sRGB approximations (low chroma, near-neutral darks):
//   bg #0d1110 · bg-2 #101413 · surface #131817 · border #1d2422
//   text #f2f5f3 · text-2 #c3cec9 · text-3 #91a09b · text-4 #697573
//   accent #4ee39a

var otpTmpl = template.Must(template.New("otp").Parse(`<!DOCTYPE html>
<html lang="fr">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Votre code de connexion trippier</title>
</head>
<body style="margin:0;padding:0;background:#0d1110;font-family:'Segoe UI',ui-sans-serif,Arial,sans-serif;-webkit-font-smoothing:antialiased;">

<table width="100%" cellpadding="0" cellspacing="0" style="background:#0d1110;padding:48px 16px 64px;">
<tr><td align="center">
<table width="100%" cellpadding="0" cellspacing="0" style="max-width:720px;">
<tr><td>

  <!-- ── Mock email client window ─────────────────────────── -->
  <table width="100%" cellpadding="0" cellspacing="0" style="background:#101413;border:1px solid #1d2422;border-radius:14px;overflow:hidden;box-shadow:0 30px 80px -30px rgba(0,0,0,0.7);">

    <!-- Window chrome / title bar -->
    <tr><td style="background:#131817;border-bottom:1px solid #1d2422;padding:12px 16px;">
      <table width="100%" cellpadding="0" cellspacing="0"><tr>
        <!-- Apple dots -->
        <td style="width:11px;padding:0;"><div style="width:11px;height:11px;background:#e78f8f;border-radius:50%;"></div></td>
        <td style="width:8px;"></td>
        <td style="width:11px;padding:0;"><div style="width:11px;height:11px;background:#e4c87a;border-radius:50%;"></div></td>
        <td style="width:8px;"></td>
        <td style="width:11px;padding:0;"><div style="width:11px;height:11px;background:#4ee39a;border-radius:50%;"></div></td>
        <!-- Title -->
        <td style="padding-left:12px;font-family:'Courier New',ui-monospace,monospace;font-size:12px;color:#91a09b;">
          inbox · {{.Email}}
        </td>
        <!-- Action icons (unicode approximations) -->
        <td style="text-align:right;font-size:13px;color:#91a09b;letter-spacing:4px;">&#9993;&nbsp;&nbsp;&#128465;&nbsp;&nbsp;&#8942;</td>
      </tr></table>
    </td></tr>

    <!-- From / To / Subject meta -->
    <tr><td style="border-bottom:1px solid #1d2422;padding:18px 24px;">
      <table width="100%" cellpadding="0" cellspacing="0">
        <tr>
          <td style="width:72px;font-family:'Courier New',ui-monospace,monospace;font-size:11px;text-transform:uppercase;letter-spacing:0.06em;color:#91a09b;padding:4px 0;">De</td>
          <td style="font-size:13px;color:#c3cec9;padding:4px 0;"><strong style="color:#f2f5f3;font-weight:500;">trippier</strong> &lt;noreply@trippier.dev&gt;</td>
        </tr>
        <tr>
          <td style="width:72px;font-family:'Courier New',ui-monospace,monospace;font-size:11px;text-transform:uppercase;letter-spacing:0.06em;color:#91a09b;padding:4px 0;">À</td>
          <td style="font-size:13px;color:#c3cec9;padding:4px 0;">{{.Email}}</td>
        </tr>
        <tr>
          <td style="width:72px;font-family:'Courier New',ui-monospace,monospace;font-size:11px;text-transform:uppercase;letter-spacing:0.06em;color:#91a09b;padding:4px 0;">Objet</td>
          <td style="font-size:13px;padding:4px 0;">
            <strong style="color:#f2f5f3;font-weight:500;">Votre code de connexion : {{.D1}}{{.D2}}{{.D3}} {{.D4}}{{.D5}}{{.D6}}</strong>
            <span style="display:inline-block;font-family:'Courier New',ui-monospace,monospace;font-size:10.5px;padding:2px 7px;border-radius:4px;color:#4ee39a;background:rgba(78,227,154,0.12);border:1px solid rgba(78,227,154,0.22);margin-left:6px;vertical-align:middle;">vérifié</span>
          </td>
        </tr>
      </table>
    </td></tr>

    <!-- ── Email body ──────────────────────────────────────── -->
    <tr><td style="padding:48px 44px 40px;background:#101413;">

      <!-- Brand -->
      <div style="font-size:15px;font-weight:600;color:#f2f5f3;letter-spacing:-0.01em;margin-bottom:28px;">
        trippier<span style="color:#697573;">/</span>api
      </div>

      <h2 style="margin:0 0 14px;font-size:26px;font-weight:600;color:#f2f5f3;line-height:1.2;letter-spacing:-0.02em;">
        Confirmez votre adresse pour vous connecter.
      </h2>
      <p style="margin:0 0 14px;font-size:15px;line-height:1.6;color:#c3cec9;">
        Quelqu'un — sûrement vous — essaie d'accéder au dashboard trippier avec
        l'adresse <span style="color:#4ee39a;">{{.Email}}</span>.
        Utilisez le code ci-dessous pour terminer la connexion.
      </p>

      <!-- OTP box -->
      <table width="100%" cellpadding="0" cellspacing="0" style="margin:32px 0 22px;">
      <tr><td style="background:#0d1110;border:1px solid #1d2422;border-left:3px solid #4ee39a;border-radius:12px;padding:28px 24px;text-align:center;">
        <div style="font-family:'Courier New',ui-monospace,monospace;font-size:10.5px;text-transform:uppercase;letter-spacing:0.08em;color:#91a09b;margin-bottom:14px;">Code à usage unique</div>
        <table cellpadding="0" cellspacing="0" style="margin:0 auto;">
        <tr>
          <td style="width:52px;height:64px;text-align:center;vertical-align:middle;font-family:'Courier New',ui-monospace,monospace;font-size:32px;font-weight:600;color:#f2f5f3;background:#131817;border:1px solid #1d2422;border-radius:8px;">{{.D1}}</td>
          <td style="width:8px;"></td>
          <td style="width:52px;height:64px;text-align:center;vertical-align:middle;font-family:'Courier New',ui-monospace,monospace;font-size:32px;font-weight:600;color:#f2f5f3;background:#131817;border:1px solid #1d2422;border-radius:8px;">{{.D2}}</td>
          <td style="width:8px;"></td>
          <td style="width:52px;height:64px;text-align:center;vertical-align:middle;font-family:'Courier New',ui-monospace,monospace;font-size:32px;font-weight:600;color:#f2f5f3;background:#131817;border:1px solid #1d2422;border-radius:8px;">{{.D3}}</td>
          <td style="width:12px;text-align:center;vertical-align:middle;font-family:'Courier New',ui-monospace,monospace;font-size:24px;color:#697573;">·</td>
          <td style="width:52px;height:64px;text-align:center;vertical-align:middle;font-family:'Courier New',ui-monospace,monospace;font-size:32px;font-weight:600;color:#f2f5f3;background:#131817;border:1px solid #1d2422;border-radius:8px;">{{.D4}}</td>
          <td style="width:8px;"></td>
          <td style="width:52px;height:64px;text-align:center;vertical-align:middle;font-family:'Courier New',ui-monospace,monospace;font-size:32px;font-weight:600;color:#f2f5f3;background:#131817;border:1px solid #1d2422;border-radius:8px;">{{.D5}}</td>
          <td style="width:8px;"></td>
          <td style="width:52px;height:64px;text-align:center;vertical-align:middle;font-family:'Courier New',ui-monospace,monospace;font-size:32px;font-weight:600;color:#f2f5f3;background:#131817;border:1px solid #1d2422;border-radius:8px;">{{.D6}}</td>
        </tr>
        </table>
        <table cellpadding="0" cellspacing="0" style="margin:18px auto 0;">
        <tr>
          <td style="width:7px;padding:0;vertical-align:middle;"><div style="width:7px;height:7px;background:#4ee39a;border-radius:50%;"></div></td>
          <td style="width:6px;"></td>
          <td style="font-family:'Courier New',ui-monospace,monospace;font-size:11.5px;color:#91a09b;vertical-align:middle;">expire dans 15 minutes</td>
        </tr>
        </table>
      </td></tr>
      </table>

      <!-- Security callout -->
      <table width="100%" cellpadding="0" cellspacing="0" style="margin-top:32px;">
      <tr><td style="background:#131817;border:1px solid #1d2422;border-left:3px solid #697573;border-radius:8px;padding:14px 16px;font-size:13px;line-height:1.55;color:#c3cec9;">
        <strong style="color:#f2f5f3;">Vous n'avez pas demandé ce code ?</strong>
        Vous pouvez ignorer ce message — personne n'accédera à votre compte sans ce code.
        Si vous recevez plusieurs tentatives, prévenez-nous à
        <a href="mailto:security@trippier.dev" style="color:#4ee39a;text-decoration:underline;">security@trippier.dev</a>.
      </td></tr>
      </table>

      <!-- Request meta -->
      <table width="100%" cellpadding="0" cellspacing="0" style="margin-top:24px;background:#0d1110;border:1px solid #1d2422;border-radius:8px;">
      <tr><td style="padding:14px 16px;">
        <table width="100%" cellpadding="0" cellspacing="0"><tr>
          <td width="50%">
            <div style="font-family:'Courier New',ui-monospace,monospace;font-size:10.5px;text-transform:uppercase;letter-spacing:0.06em;color:#697573;margin-bottom:3px;">demandé</div>
            <div style="font-family:'Courier New',ui-monospace,monospace;font-size:12px;color:#f2f5f3;">{{.SentAt}}</div>
          </td>
          <td width="50%">
            <div style="font-family:'Courier New',ui-monospace,monospace;font-size:10.5px;text-transform:uppercase;letter-spacing:0.06em;color:#697573;margin-bottom:3px;">ip</div>
            <div style="font-family:'Courier New',ui-monospace,monospace;font-size:12px;color:#f2f5f3;">{{.IP}}</div>
          </td>
        </tr></table>
      </td></tr>
      </table>

      <!-- Footer -->
      <table width="100%" cellpadding="0" cellspacing="0" style="margin-top:36px;border-top:1px solid #1d2422;">
      <tr><td style="padding-top:24px;text-align:center;color:#697573;font-size:12px;line-height:1.6;">
        trippier/api est open source —
        <a href="https://github.com/ulysse-mercadal/trippier-public-API" style="color:#91a09b;text-decoration:underline;">github.com/ulysse-mercadal/trippier-public-API</a><br>
        Vous recevez cet email parce qu'une connexion a été tentée avec cette adresse.
      </td></tr>
      </table>

    </td></tr>
    <!-- ── /Email body ─────────────────────────────────────── -->

  </table>

</td></tr>
</table>
</td></tr>
</table>

</body>
</html>`))

type otpData struct {
	Code, Email, SentAt, ExpiresAt string
	IP                             string
	D1, D2, D3, D4, D5, D6         string
}

// SendOTPCode sends a 6-digit verification code to addr.
func (s *Sender) SendOTPCode(to, code, clientIP, userAgent, appURL string) error {
	digits := []rune(code)
	for len(digits) < 6 {
		digits = append(digits, '0')
	}
	now := time.Now().UTC()
	if clientIP == "" {
		clientIP = "—"
	}
	data := otpData{
		Code:      code,
		Email:     to,
		SentAt:    now.Format("2 Jan 2006 · 15:04 UTC"),
		ExpiresAt: now.Add(15 * time.Minute).Format("15:04 UTC"),
		IP:        clientIP,
		D1:        string(digits[0]), D2: string(digits[1]), D3: string(digits[2]),
		D4: string(digits[3]), D5: string(digits[4]), D6: string(digits[5]),
	}
	var buf bytes.Buffer
	if err := otpTmpl.Execute(&buf, data); err != nil {
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

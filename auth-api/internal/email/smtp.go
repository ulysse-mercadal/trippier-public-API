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
<title>Email OTP — trippier/api</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Space+Grotesk:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500;600;700&display=swap" rel="stylesheet">
<style>
  :root {
    --bg:        oklch(13% 0.012 175);
    --bg-2:      oklch(15% 0.013 175);
    --surface:   oklch(17% 0.014 175);
    --surface-2: oklch(20% 0.015 175);
    --border:    oklch(26% 0.018 175);
    --border-2:  oklch(32% 0.02 175);
    --text:      oklch(96% 0.005 175);
    --text-2:    oklch(78% 0.008 175);
    --text-3:    oklch(60% 0.012 175);
    --text-4:    oklch(45% 0.012 175);
    --accent:    #34d39c;
    --accent-soft: #34d39c22;
    --font-sans: 'Space Grotesk', ui-sans-serif, sans-serif;
    --font-mono: 'JetBrains Mono', ui-monospace, monospace;
  }
  * { box-sizing: border-box; }
  html, body {
    margin: 0;
    padding: 0;
    background: var(--bg);
    color: var(--text);
    font-family: var(--font-sans);
    -webkit-font-smoothing: antialiased;
    min-height: 100vh;
  }
  .stage {
    min-height: 100vh;
    display: flex;
    align-items: flex-start;
    justify-content: center;
    padding: 56px 32px 80px;
    position: relative;
  }
  .topo {
    position: fixed;
    inset: 0;
    z-index: 0;
    pointer-events: none;
    background:
      radial-gradient(80% 60% at 20% 0%,  color-mix(in oklch, var(--accent) 6%, transparent), transparent 60%),
      radial-gradient(70% 50% at 100% 30%, color-mix(in oklch, var(--accent) 4%, transparent), transparent 60%);
  }
  .topo svg { width: 100%; height: 100%; }
  .col {
    position: relative;
    z-index: 1;
    display: flex;
    flex-direction: column;
    gap: 24px;
    align-items: center;
    width: 100%;
  }
  .client {
    width: 100%;
    max-width: 720px;
    background: var(--bg-2);
    border: 1px solid var(--border);
    border-radius: 14px;
    overflow: hidden;
    box-shadow:
      0 30px 80px -30px rgba(0,0,0,0.7),
      0 0 0 1px color-mix(in oklch, var(--accent) 4%, transparent);
  }
  .client-bar {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 12px 16px;
    background: var(--surface);
    border-bottom: 1px solid var(--border);
  }
  .client-bar .dot { width: 11px; height: 11px; border-radius: 50%; background: var(--surface-2); }
  .client-bar .dot.r { background: #e78f8f; }
  .client-bar .dot.y { background: #e4c87a; }
  .client-bar .dot.g { background: var(--accent); }
  .client-bar .title { margin-left: 12px; font-family: var(--font-mono); font-size: 12px; color: var(--text-3); flex: 1; }
  .client-bar .actions { display: flex; gap: 8px; color: var(--text-3); }
  .client-bar .actions svg { display: block; }
  .meta {
    padding: 18px 24px;
    border-bottom: 1px solid var(--border);
    display: grid;
    grid-template-columns: 72px 1fr;
    gap: 8px 16px;
    font-size: 13px;
    line-height: 1.5;
  }
  .meta .k { font-family: var(--font-mono); font-size: 11px; text-transform: uppercase; letter-spacing: 0.06em; color: var(--text-3); padding-top: 2px; }
  .meta .v { color: var(--text-2); }
  .meta .v strong { color: var(--text); font-weight: 500; }
  .meta .v .pill {
    display: inline-block;
    font-family: var(--font-mono);
    font-size: 10.5px;
    padding: 2px 7px;
    border-radius: 4px;
    color: var(--accent);
    background: color-mix(in oklch, var(--accent) 12%, transparent);
    border: 1px solid color-mix(in oklch, var(--accent) 22%, transparent);
    margin-left: 6px;
    vertical-align: middle;
  }
  .email {
    padding: 48px 56px 40px;
    background:
      radial-gradient(80% 50% at 50% 0%, color-mix(in oklch, var(--accent) 4%, transparent), transparent 60%),
      var(--bg-2);
    position: relative;
  }
  .brand {
    display: inline-flex;
    align-items: center;
    gap: 10px;
    font-weight: 600;
    font-size: 15px;
    color: var(--text);
    letter-spacing: -0.01em;
    margin-bottom: 28px;
  }
  .brand svg { color: var(--accent); }
  .brand .dot { color: var(--text-3); margin: 0 1px; }
  .email h2 { font-size: 26px; font-weight: 600; letter-spacing: -0.02em; line-height: 1.2; margin: 0 0 14px; color: var(--text); }
  .email p { font-size: 15px; line-height: 1.6; color: var(--text-2); margin: 0 0 14px; }
  .email p .accent { color: var(--accent); }
  .otp-wrap {
    margin: 32px 0 22px;
    padding: 28px 24px;
    background: var(--bg);
    border: 1px solid var(--border);
    border-radius: 12px;
    text-align: center;
    position: relative;
    border-left: 3px solid var(--accent);
  }
  .otp-label { font-family: var(--font-mono); font-size: 11px; text-transform: uppercase; letter-spacing: 0.08em; color: var(--text-3); margin-bottom: 14px; }
  .otp-code { display: inline-flex; gap: 8px; font-family: var(--font-mono); font-weight: 600; color: var(--text); user-select: all; cursor: text; }
  .otp-digit {
    width: 52px; height: 64px;
    display: inline-flex; align-items: center; justify-content: center;
    font-size: 32px;
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 8px;
    letter-spacing: 0;
  }
  .otp-digit.sep { width: 12px; background: none; border: none; color: var(--text-3); font-size: 24px; }
  .otp-expiry {
    margin-top: 18px;
    font-family: var(--font-mono);
    font-size: 11.5px;
    color: var(--text-3);
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }
  .otp-expiry::before {
    content: '';
    width: 7px; height: 7px;
    border-radius: 50%;
    background: var(--accent);
    box-shadow: 0 0 0 3px color-mix(in oklch, var(--accent) 22%, transparent);
  }
  .security {
    margin-top: 32px;
    padding: 14px 16px;
    border: 1px solid var(--border);
    border-left: 3px solid var(--text-3);
    border-radius: 8px;
    background: color-mix(in oklch, var(--surface) 40%, transparent);
    display: flex;
    gap: 12px;
    font-size: 13px;
    line-height: 1.55;
    color: var(--text-2);
  }
  .security svg { color: var(--text-3); flex-shrink: 0; margin-top: 1px; }
  .security strong { color: var(--text); margin-right: 4px; }
  .email-foot {
    margin-top: 36px;
    padding-top: 24px;
    border-top: 1px solid var(--border);
    text-align: center;
    color: var(--text-4);
    font-size: 12px;
    line-height: 1.6;
  }
  .email-foot a { color: var(--text-3); text-decoration: underline; text-underline-offset: 3px; }
  @media (max-width: 640px) {
    .stage { padding: 24px 16px 48px; }
    .email { padding: 32px 28px; }
    .otp-digit { width: 38px; height: 52px; font-size: 24px; }
    .meta { grid-template-columns: 60px 1fr; padding: 14px 18px; }
  }
</style>
</head>
<body>

<div class="topo" aria-hidden="true">
  <svg viewBox="0 0 1600 1100" preserveAspectRatio="xMidYMid slice">
    <defs>
      <radialGradient id="fade" cx="50%" cy="35%" r="80%">
        <stop offset="0%" stop-color="white" stop-opacity="1"/>
        <stop offset="70%" stop-color="white" stop-opacity="0.4"/>
        <stop offset="100%" stop-color="white" stop-opacity="0"/>
      </radialGradient>
      <mask id="topo-fade-m"><rect width="100%" height="100%" fill="url(#fade)"/></mask>
    </defs>
    <g mask="url(#topo-fade-m)" fill="none" stroke="#34d39c" stroke-width="0.6" opacity="0.12">
      <path d="M-50 120 Q 200 60 450 130 T 950 100 T 1450 150 T 1750 90"/>
      <path d="M-50 200 Q 220 140 470 210 T 970 180 T 1470 230 T 1750 170"/>
      <path d="M-50 280 Q 240 220 490 290 T 990 260 T 1490 310 T 1750 250"/>
      <path d="M-50 360 Q 260 300 510 370 T 1010 340 T 1510 390 T 1750 330"/>
      <path d="M-50 440 Q 280 380 530 450 T 1030 420 T 1530 470 T 1750 410"/>
      <path d="M-50 540 Q 240 480 490 550 T 990 520 T 1490 570 T 1750 510"/>
      <path d="M-50 640 Q 220 580 470 650 T 970 620 T 1470 670 T 1750 610"/>
      <path d="M-50 740 Q 260 680 510 750 T 1010 720 T 1510 770 T 1750 710"/>
      <path d="M-50 840 Q 280 780 530 850 T 1030 820 T 1530 870 T 1750 810"/>
      <path d="M-50 940 Q 300 880 550 950 T 1050 920 T 1550 970 T 1750 910"/>
    </g>
    <g mask="url(#topo-fade-m)" fill="none" stroke="#34d39c" stroke-width="1" opacity="0.18">
      <path d="M-50 400 Q 280 340 530 410 T 1030 380 T 1530 430 T 1750 370"/>
      <path d="M-50 700 Q 260 640 510 710 T 1010 680 T 1510 730 T 1750 670"/>
    </g>
  </svg>
</div>

<div class="stage">
  <div class="col">
    <div class="client">
      <div class="client-bar">
        <span class="dot r"></span>
        <span class="dot y"></span>
        <span class="dot g"></span>
        <span class="title">inbox · {{.Email}}</span>
        <span class="actions">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M3 7l9 6 9-6M3 7v10a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V7M3 7l2-2h14l2 2"/></svg>
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M4 7h16M9 7V5a2 2 0 0 1 2-2h2a2 2 0 0 1 2 2v2M6 7v13a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2V7"/></svg>
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="6" r="1.5"/><circle cx="12" cy="12" r="1.5"/><circle cx="12" cy="18" r="1.5"/></svg>
        </span>
      </div>

      <div class="meta">
        <div class="k">De</div>
        <div class="v"><strong>trippier</strong> &lt;noreply@trippier.dev&gt;</div>
        <div class="k">À</div>
        <div class="v">{{.Email}}</div>
        <div class="k">Objet</div>
        <div class="v"><strong>Votre code de connexion : {{.Code}}</strong> <span class="pill">vérifié</span></div>
      </div>

      <div class="email">
        <div class="brand">
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
            <path d="M4 18c4-8 12-8 16 0"/>
            <path d="M6 14c3-5 9-5 12 0" opacity="0.6"/>
            <path d="M9 10c1.5-2 4.5-2 6 0" opacity="0.3"/>
            <circle cx="12" cy="20" r="1.2" fill="currentColor"/>
          </svg>
          <span>trippier<span class="dot">/</span>api</span>
        </div>

        <h2>Confirmez votre adresse pour vous connecter.</h2>
        <p>
          Quelqu'un — sûrement vous — essaie d'accéder au dashboard trippier avec
          l'adresse <span class="accent">{{.Email}}</span>. Utilisez le code ci-dessous pour terminer la connexion.
        </p>

        <div class="otp-wrap">
          <div class="otp-label">Code à usage unique</div>
          <div class="otp-code" aria-label="Code OTP : {{.Code}}">
            <span class="otp-digit">{{printf "%c" (index .Code 0)}}</span>
            <span class="otp-digit">{{printf "%c" (index .Code 1)}}</span>
            <span class="otp-digit">{{printf "%c" (index .Code 2)}}</span>
            <span class="otp-digit sep">·</span>
            <span class="otp-digit">{{printf "%c" (index .Code 3)}}</span>
            <span class="otp-digit">{{printf "%c" (index .Code 4)}}</span>
            <span class="otp-digit">{{printf "%c" (index .Code 5)}}</span>
          </div>
          <div class="otp-expiry">expire dans 15 minutes</div>
        </div>

        <div class="security">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M12 2 4 6v6c0 5 3.5 8.5 8 10 4.5-1.5 8-5 8-10V6l-8-4z"/>
            <path d="m9 12 2 2 4-4"/>
          </svg>
          <div>
            <strong>Vous n'avez pas demandé ce code ?</strong>
            Vous pouvez ignorer ce message — personne n'accédera à votre compte sans ce code.
          </div>
        </div>

        <div class="email-foot">
          trippier/api est open source — <a href="https://github.com/ulysse-mercadal/trippier-public-API">github.com/ulysse-mercadal/trippier-public-API</a><br>
          Vous recevez cet email parce qu'une connexion a été tentée avec cette adresse.
        </div>
      </div>
    </div>
  </div>
</div>

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

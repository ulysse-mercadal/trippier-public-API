package email_test

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/trippier/auth-api/internal/email"
)

// fakeSMTP is a minimal SMTP server for tests.
// It handles a single connection, captures the DATA payload, and responds
// with the minimum set of codes that net/smtp.SendMail expects.
type fakeSMTP struct {
	ln   net.Listener
	mu   sync.Mutex
	body string
}

func startFakeSMTP(t *testing.T) *fakeSMTP {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	f := &fakeSMTP{ln: ln}
	t.Cleanup(func() { ln.Close() })
	go f.serve()
	return f
}

func (f *fakeSMTP) host() string {
	host, _, _ := net.SplitHostPort(f.ln.Addr().String())
	return host
}

func (f *fakeSMTP) port() int {
	_, p, _ := net.SplitHostPort(f.ln.Addr().String())
	n, _ := strconv.Atoi(p)
	return n
}

func (f *fakeSMTP) captured() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.body
}

func (f *fakeSMTP) serve() {
	conn, err := f.ln.Accept()
	if err != nil {
		return
	}
	defer conn.Close()

	r := bufio.NewReader(conn)

	writeLine := func(s string) {
		fmt.Fprintf(conn, "%s\r\n", s)
	}
	readLine := func() string {
		line, _ := r.ReadString('\n')
		return strings.TrimRight(line, "\r\n")
	}

	writeLine("220 test SMTP server")

	inData := false
	var dataBuf strings.Builder

	for {
		line := readLine()
		if line == "" {
			continue
		}
		upper := strings.ToUpper(line)

		if inData {
			if line == "." {
				f.mu.Lock()
				f.body = dataBuf.String()
				f.mu.Unlock()
				writeLine("250 Message accepted")
				inData = false
			} else {
				// Dot-stuffing: a leading dot that isn't the terminator
				line = strings.TrimPrefix(line, ".")
				dataBuf.WriteString(line + "\n")
			}
		} else {
			switch {
			case strings.HasPrefix(upper, "EHLO"), strings.HasPrefix(upper, "HELO"):
				writeLine("250 Hello")
			case strings.HasPrefix(upper, "MAIL FROM"), strings.HasPrefix(upper, "RCPT TO"):
				writeLine("250 OK")
			case upper == "DATA":
				writeLine("354 Start input; end with <CRLF>.<CRLF>")
				inData = true
				dataBuf.Reset()
			case strings.HasPrefix(upper, "QUIT"):
				writeLine("221 Bye")
				return
			default:
				writeLine("250 OK")
			}
		}
	}
}

func TestSendVerification_BodyContainsURL(t *testing.T) {
	srv := startFakeSMTP(t)

	s := email.New(srv.host(), srv.port(), "noreply@trippier.dev")
	verifyURL := "https://trippier.dev/api/auth/verify-email?token=abc123test456"

	if err := s.SendVerification("user@example.com", verifyURL); err != nil {
		t.Fatalf("SendVerification: %v", err)
	}

	body := srv.captured()
	if !strings.Contains(body, verifyURL) {
		t.Errorf("email body should contain the verify URL\nbody:\n%s", body)
	}
}

func TestSendVerification_BodyContainsSubject(t *testing.T) {
	srv := startFakeSMTP(t)

	s := email.New(srv.host(), srv.port(), "noreply@trippier.dev")
	if err := s.SendVerification("user@example.com", "https://example.com/verify"); err != nil {
		t.Fatalf("SendVerification: %v", err)
	}

	body := srv.captured()
	if !strings.Contains(body, "Confirm your Trippier account") {
		t.Errorf("email body should contain subject line\nbody:\n%s", body)
	}
}

func TestSendVerification_URLAppearsInHTMLLink(t *testing.T) {
	srv := startFakeSMTP(t)

	s := email.New(srv.host(), srv.port(), "noreply@trippier.dev")
	verifyURL := "https://trippier.dev/verify?token=unique-token-xyz"

	if err := s.SendVerification("to@example.com", verifyURL); err != nil {
		t.Fatalf("SendVerification: %v", err)
	}

	body := srv.captured()
	// URL should appear at least twice: once in the <a href> and once as fallback text.
	count := strings.Count(body, verifyURL)
	if count < 2 {
		t.Errorf("expected URL to appear at least twice in body (href + fallback), got %d\nbody:\n%s", count, body)
	}
}

func TestNew_SMTP_Unreachable(t *testing.T) {
	s := email.New("127.0.0.1", 1, "noreply@trippier.dev")
	err := s.SendVerification("to@example.com", "https://example.com/verify")
	if err == nil {
		t.Error("expected error when SMTP host is unreachable, got nil")
	}
}

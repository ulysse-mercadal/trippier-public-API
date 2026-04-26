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

func TestSendOTPCode_BodyContainsCode(t *testing.T) {
	srv := startFakeSMTP(t)

	s := email.New(srv.host(), srv.port(), "noreply@trippier.dev")
	code := "482917"

	if err := s.SendOTPCode("user@example.com", code); err != nil {
		t.Fatalf("SendOTPCode: %v", err)
	}

	body := srv.captured()
	if !strings.Contains(body, code) {
		t.Errorf("email body should contain the OTP code\nbody:\n%s", body)
	}
}

func TestSendOTPCode_BodyContainsSubject(t *testing.T) {
	srv := startFakeSMTP(t)

	s := email.New(srv.host(), srv.port(), "noreply@trippier.dev")
	if err := s.SendOTPCode("user@example.com", "123456"); err != nil {
		t.Fatalf("SendOTPCode: %v", err)
	}

	body := srv.captured()
	if !strings.Contains(body, "verification code") {
		t.Errorf("email body should contain 'verification code'\nbody:\n%s", body)
	}
}

func TestSendOTPCode_SubjectContainsCode(t *testing.T) {
	srv := startFakeSMTP(t)

	s := email.New(srv.host(), srv.port(), "noreply@trippier.dev")
	code := "739201"

	if err := s.SendOTPCode("to@example.com", code); err != nil {
		t.Fatalf("SendOTPCode: %v", err)
	}

	body := srv.captured()
	// The Subject header should include the code for email preview snippets.
	if !strings.Contains(body, "Subject:") || !strings.Contains(body, code) {
		t.Errorf("email should contain Subject header with code\nbody:\n%s", body)
	}
}

func TestNew_SMTP_Unreachable(t *testing.T) {
	s := email.New("127.0.0.1", 1, "noreply@trippier.dev")
	err := s.SendOTPCode("to@example.com", "000000")
	if err == nil {
		t.Error("expected error when SMTP host is unreachable, got nil")
	}
}

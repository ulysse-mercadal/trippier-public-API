package providers

import "testing"

// TestNormalizeLang checks that valid Wikimedia codes pass through (lowercased)
// and that SSRF-prone or malformed inputs collapse to the empty fallback.
func TestNormalizeLang(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"en", "en"},
		{"FR", "fr"},
		{"  de ", "de"},
		{"zh-yue", "zh-yue"},
		{"bat-smg", "bat-smg"},
		{"", ""},
		{"evil.com", ""},
		{"en/../x", ""},
		{"en:8080", ""},
		{"a_b", ""},
		{"toolonglanguagecode", ""},
	}
	for _, c := range cases {
		if got := NormalizeLang(c.in); got != c.want {
			t.Errorf("NormalizeLang(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

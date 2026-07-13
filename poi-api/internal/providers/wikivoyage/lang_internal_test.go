package wikivoyage

import "testing"

// TestForLangSwapsEdition verifies that a per-request language retargets the
// MediaWiki endpoint, that invalid/empty codes keep the default edition, and
// that a pinned (test) endpoint is never rewritten.
func TestForLangSwapsEdition(t *testing.T) {
	p := New("en")

	if got := p.forLang("fr").baseURL; got != "https://fr.wikivoyage.org/w/api.php" {
		t.Errorf("forLang(fr) baseURL = %q", got)
	}
	if got := p.forLang("").baseURL; got != "https://en.wikivoyage.org/w/api.php" {
		t.Errorf("forLang(\"\") should keep default, got %q", got)
	}
	if got := p.forLang("../evil").baseURL; got != "https://en.wikivoyage.org/w/api.php" {
		t.Errorf("forLang(malicious) should keep default, got %q", got)
	}

	tp := NewWithURL("http://127.0.0.1:1/api")
	if got := tp.forLang("fr").baseURL; got != "http://127.0.0.1:1/api" {
		t.Errorf("pinned test URL must not swap, got %q", got)
	}
}

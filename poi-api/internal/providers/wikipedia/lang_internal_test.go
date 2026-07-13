package wikipedia

import "testing"

// TestForLangSwapsEdition verifies that a per-request language retargets the
// MediaWiki endpoint, that invalid/empty codes keep the default edition, and
// that a pinned (test) endpoint is never rewritten.
func TestForLangSwapsEdition(t *testing.T) {
	p := New("en")

	if got := p.base.forLang("fr").baseURL; got != "https://fr.wikipedia.org/w/api.php" {
		t.Errorf("forLang(fr) baseURL = %q", got)
	}
	if got := p.base.forLang("").baseURL; got != "https://en.wikipedia.org/w/api.php" {
		t.Errorf("forLang(\"\") should keep default, got %q", got)
	}
	if got := p.base.forLang("evil.com/x").baseURL; got != "https://en.wikipedia.org/w/api.php" {
		t.Errorf("forLang(malicious) should keep default, got %q", got)
	}

	tp := NewWithURLs("http://127.0.0.1:1/api", "http://127.0.0.1:1/sparql")
	if got := tp.base.forLang("fr").baseURL; got != "http://127.0.0.1:1/api" {
		t.Errorf("pinned test URL must not swap, got %q", got)
	}
}

package wikivoyage

import "testing"

func TestWikipediaURL(t *testing.T) {
	en := New("en")
	fr := New("fr")
	cases := []struct {
		desc string
		p    *Provider
		in   string
		want string
	}{
		{"empty", en, "", ""},
		{"plain title defaults to en for en.wikivoyage", en, "Eiffel Tower", "https://en.wikipedia.org/wiki/Eiffel_Tower"},
		{"plain title defaults to fr for fr.wikivoyage", fr, "Tour Eiffel", "https://fr.wikipedia.org/wiki/Tour_Eiffel"},
		{"interwiki prefix overrides default lang", en, "fr:Tour Eiffel", "https://fr.wikipedia.org/wiki/Tour_Eiffel"},
		{"full URL is returned as-is", en, "https://de.wikipedia.org/wiki/Eiffelturm", "https://de.wikipedia.org/wiki/Eiffelturm"},
		{"plain http URL is returned as-is", en, "http://en.wikipedia.org/wiki/Foo", "http://en.wikipedia.org/wiki/Foo"},
		{"whitespace only is empty", en, "   ", ""},
		{"title with special chars is escaped", en, "Café de Flore", "https://en.wikipedia.org/wiki/Caf%C3%A9_de_Flore"},
		{"interwiki prefix with empty title yields empty", en, "fr:", ""},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			if got := c.p.wikipediaURL(c.in); got != c.want {
				t.Errorf("wikipediaURL(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestLangCode(t *testing.T) {
	cases := []struct {
		baseURL string
		want    string
	}{
		{"https://en.wikivoyage.org/w/api.php", "en"},
		{"https://fr.wikivoyage.org/w/api.php", "fr"},
		{"https://de.wikivoyage.org/w/api.php", "de"},
		{"", "en"},
		{"not a url at all", "en"},
	}
	for _, c := range cases {
		p := &Provider{baseURL: c.baseURL}
		if got := p.langCode(); got != c.want {
			t.Errorf("langCode for %q = %q, want %q", c.baseURL, got, c.want)
		}
	}
}

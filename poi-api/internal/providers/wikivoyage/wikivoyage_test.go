package wikivoyage_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/trippier/poi-api/internal/providers/wikivoyage"
	"github.com/trippier/poi-api/pkg/types"
)

// newServer builds a test server that routes on the "action" query parameter.
func newServer(t *testing.T, geosearchTitle string, wikitext string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		action := r.URL.Query().Get("action")
		w.Header().Set("Content-Type", "application/json")
		switch action {
		case "query":
			json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
				"query": map[string]any{
					"geosearch": []map[string]any{
						{"title": geosearchTitle},
					},
				},
			})
		case "parse":
			json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
				"parse": map[string]any{
					"wikitext": map[string]any{
						"*": wikitext,
					},
				},
			})
		default:
			http.Error(w, "unknown action", http.StatusBadRequest)
		}
	}))
}

const sampleWikitext = `
{{see|name=Eiffel Tower|lat=48.8584|long=2.2945|url=https://toureiffel.paris|content=Iconic iron lattice tower.}}
{{eat|name=Le Jules Verne|lat=48.8583|long=2.2944|hours=12:00-22:00}}
{{see|name=No Coords Place|content=Somewhere vague}}
{{see|}}
`

// ── SupportsMode ─────────────────────────────────────────────────────────────

func TestSupportsMode(t *testing.T) {
	p := wikivoyage.New("en")
	cases := []struct {
		mode types.SearchMode
		want bool
	}{
		{types.ModeRadius, true},
		{types.ModeDistrict, true},
		{types.ModePolygon, false},
	}
	for _, tc := range cases {
		got := p.SupportsMode(tc.mode)
		if got != tc.want {
			t.Errorf("SupportsMode(%s) = %v, want %v", tc.mode, got, tc.want)
		}
	}
}

// ── District mode (single API call) ──────────────────────────────────────────

func TestSearch_DistrictMode_ParsesListings(t *testing.T) {
	srv := newServer(t, "Paris", sampleWikitext)
	defer srv.Close()

	p := wikivoyage.NewWithURL(srv.URL)
	pois, err := p.Search(context.Background(), types.SearchQuery{
		Mode:     types.ModeDistrict,
		District: "Paris",
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	// 4 templates: Eiffel Tower, Le Jules Verne, No Coords Place — unnamed one skipped.
	if len(pois) != 3 {
		t.Fatalf("expected 3 POIs (unnamed template skipped), got %d", len(pois))
	}
}

func TestSearch_DistrictMode_TypeMapping(t *testing.T) {
	srv := newServer(t, "Paris", sampleWikitext)
	defer srv.Close()

	p := wikivoyage.NewWithURL(srv.URL)
	pois, err := p.Search(context.Background(), types.SearchQuery{
		Mode:     types.ModeDistrict,
		District: "Paris",
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	byName := map[string]types.RawPoi{}
	for _, poi := range pois {
		byName[poi.Name] = poi
	}

	if byName["Eiffel Tower"].Type != types.TypeSee {
		t.Errorf("Eiffel Tower: want TypeSee, got %s", byName["Eiffel Tower"].Type)
	}
	if byName["Le Jules Verne"].Type != types.TypeEat {
		t.Errorf("Le Jules Verne: want TypeEat, got %s", byName["Le Jules Verne"].Type)
	}
}

func TestSearch_DistrictMode_SourceURL(t *testing.T) {
	srv := newServer(t, "Paris", sampleWikitext)
	defer srv.Close()

	p := wikivoyage.NewWithURL(srv.URL + "/w/api.php")
	pois, err := p.Search(context.Background(), types.SearchQuery{
		Mode:     types.ModeDistrict,
		District: "Paris",
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	wantByName := map[string]string{
		"Eiffel Tower":    srv.URL + "/wiki/Paris#Eiffel_Tower",
		"Le Jules Verne":  srv.URL + "/wiki/Paris#Le_Jules_Verne",
		"No Coords Place": srv.URL + "/wiki/Paris#No_Coords_Place",
	}
	for _, poi := range pois {
		want, ok := wantByName[poi.Name]
		if !ok {
			t.Errorf("unexpected POI %q", poi.Name)
			continue
		}
		if poi.SourceURL != want {
			t.Errorf("%s SourceURL = %q, want %q", poi.Name, poi.SourceURL, want)
		}
	}
}

func TestSearch_DistrictMode_CoordinatesPresent(t *testing.T) {
	srv := newServer(t, "Paris", sampleWikitext)
	defer srv.Close()

	p := wikivoyage.NewWithURL(srv.URL)
	pois, err := p.Search(context.Background(), types.SearchQuery{
		Mode:     types.ModeDistrict,
		District: "Paris",
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	byName := map[string]types.RawPoi{}
	for _, poi := range pois {
		byName[poi.Name] = poi
	}

	if byName["Eiffel Tower"].Coords == nil || byName["Eiffel Tower"].Coords.Approximate {
		t.Error("Eiffel Tower should have exact coordinates")
	}
	if byName["No Coords Place"].Coords == nil || !byName["No Coords Place"].Coords.Approximate {
		t.Error("No Coords Place should have approximate=true when lat/long are missing")
	}
}

func TestSearch_DistrictMode_ContactFields(t *testing.T) {
	srv := newServer(t, "Paris", sampleWikitext)
	defer srv.Close()

	p := wikivoyage.NewWithURL(srv.URL)
	pois, err := p.Search(context.Background(), types.SearchQuery{
		Mode:     types.ModeDistrict,
		District: "Paris",
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	byName := map[string]types.RawPoi{}
	for _, poi := range pois {
		byName[poi.Name] = poi
	}

	if byName["Eiffel Tower"].Contact.Website != "https://toureiffel.paris" {
		t.Errorf("Eiffel Tower website = %q, want https://toureiffel.paris", byName["Eiffel Tower"].Contact.Website)
	}
	if byName["Le Jules Verne"].Contact.Hours != "12:00-22:00" {
		t.Errorf("Le Jules Verne hours = %q, want 12:00-22:00", byName["Le Jules Verne"].Contact.Hours)
	}
}

// ── Radius mode (two API calls: geosearch + wikitext) ─────────────────────────

func TestSearch_RadiusMode_ResolvesZoneThenFetches(t *testing.T) {
	srv := newServer(t, "Paris/7th arrondissement", sampleWikitext)
	defer srv.Close()

	p := wikivoyage.NewWithURL(srv.URL)
	pois, err := p.Search(context.Background(), types.SearchQuery{
		Mode:   types.ModeRadius,
		Lat:    48.858,
		Lng:    2.294,
		Radius: 1000,
	})
	if err != nil {
		t.Fatalf("Search radius: %v", err)
	}
	if len(pois) == 0 {
		t.Error("expected at least one POI from radius search")
	}
}

// ── Error handling ────────────────────────────────────────────────────────────

func TestSearch_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	p := wikivoyage.NewWithURL(srv.URL)
	_, err := p.Search(context.Background(), types.SearchQuery{
		Mode:     types.ModeDistrict,
		District: "Nowhere",
	})
	if err == nil {
		t.Error("expected error on HTTP 500, got nil")
	}
}

func TestSearch_EmptyWikitext_ReturnsEmpty(t *testing.T) {
	srv := newServer(t, "Empty Zone", "")
	defer srv.Close()

	p := wikivoyage.NewWithURL(srv.URL)
	pois, err := p.Search(context.Background(), types.SearchQuery{
		Mode:     types.ModeDistrict,
		District: "Empty Zone",
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(pois) != 0 {
		t.Errorf("expected 0 POIs for empty wikitext, got %d", len(pois))
	}
}

func TestSearch_PolygonMode_ReturnsNil(t *testing.T) {
	p := wikivoyage.New("en")
	pois, err := p.Search(context.Background(), types.SearchQuery{
		Mode:    types.ModePolygon,
		Polygon: "48.84 2.34 48.86 2.34",
	})
	if err != nil {
		t.Fatalf("unexpected error for unsupported mode: %v", err)
	}
	if len(pois) != 0 {
		t.Errorf("expected nil/empty for unsupported mode, got %d POIs", len(pois))
	}
}

// TestSearch_ImageFromListingField confirms the image= listing field is
// resolved to a Commons Special:FilePath URL. A listing without image= ends
// up with a nil Images slice.
func TestSearch_ImageFromListingField(t *testing.T) {
	wikitext := `
{{see|name=With Image|lat=48|long=2|image=Eiffel Tower.jpg}}
{{see|name=No Image|lat=48|long=2}}
`
	srv := newServer(t, "Zone", wikitext)
	defer srv.Close()

	p := wikivoyage.NewWithURL(srv.URL)
	pois, err := p.Search(context.Background(), types.SearchQuery{Mode: types.ModeDistrict, District: "Zone"})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	byName := map[string][]string{}
	for _, poi := range pois {
		byName[poi.Name] = poi.Images
	}

	want := "https://commons.wikimedia.org/wiki/Special:FilePath/Eiffel%20Tower.jpg"
	if got := byName["With Image"]; len(got) != 1 || got[0] != want {
		t.Errorf("With Image Images = %v, want [%q]", got, want)
	}
	if got := byName["No Image"]; len(got) != 0 {
		t.Errorf("No Image Images = %v, want nil/empty", got)
	}
}

// TestSearch_EmitsWikipediaCrossLink confirms that a listing carrying a
// wikipedia= field produces a second RawPoi with Provider=wikipedia and the
// constructed article URL — letting dedup fold the link into the merged POI
// when the user has Wikipedia among their selected providers. A listing
// without wikipedia= produces only the main Wikivoyage POI.
func TestSearch_EmitsWikipediaCrossLink(t *testing.T) {
	wikitext := `
{{see|name=Plain Name|lat=48|long=2|wikipedia=Eiffel Tower}}
{{see|name=With Lang Prefix|lat=48|long=2|wikipedia=fr:Tour Eiffel}}
{{see|name=Full URL|lat=48|long=2|wikipedia=https://de.wikipedia.org/wiki/Eiffelturm}}
{{see|name=No Wiki|lat=48|long=2}}
`
	srv := newServer(t, "Zone", wikitext)
	defer srv.Close()

	p := wikivoyage.NewWithURL(srv.URL + "/w/api.php")
	pois, err := p.Search(context.Background(), types.SearchQuery{Mode: types.ModeDistrict, District: "Zone"})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	bySourceURL := map[string]types.RawPoi{}
	for _, poi := range pois {
		if poi.Provider == types.ProviderWikipedia {
			bySourceURL[poi.SourceURL] = poi
		}
	}

	// langCode derives from baseURL host. NewWithURL uses srv.URL which is
	// 127.0.0.1:port — the lang ends up as "127" which is fine for the test
	// scaffolding; only the host-prefix cases need to match precisely.
	wantURLs := []string{
		"https://de.wikipedia.org/wiki/Eiffelturm",
		"https://fr.wikipedia.org/wiki/Tour_Eiffel",
	}
	for _, want := range wantURLs {
		if _, ok := bySourceURL[want]; !ok {
			t.Errorf("expected synthetic Wikipedia POI with SourceURL %q, got: %v", want, bySourceURL)
		}
	}

	// 4 main wikivoyage POIs + 3 synthetic wikipedia POIs (No Wiki has no field) = 7.
	if len(pois) != 7 {
		t.Errorf("expected 7 POIs (4 wikivoyage + 3 wikipedia), got %d", len(pois))
	}
	wikipediaCount := 0
	for _, poi := range pois {
		if poi.Provider == types.ProviderWikipedia {
			wikipediaCount++
		}
	}
	if wikipediaCount != 3 {
		t.Errorf("expected 3 wikipedia POIs, got %d", wikipediaCount)
	}
}

// TestSearch_NestedTemplateDoesNotTruncateListing is a regression test for
// the real Musée Jacquemart-André listing on en.wikivoyage.org: its
// directions= field uses an inner {{station|…}} template. Before scanListings
// became brace-depth-aware the outer listing terminated at the first }} from
// the nested template, dropping every field after directions (including
// wikipedia=, image=, wikidata= and content=).
func TestSearch_NestedTemplateDoesNotTruncateListing(t *testing.T) {
	wikitext := `{{see
| name=Musée Jacquemart-André | alt=Jacquemart-Andre Museum | url=http://www.musee-jacquemart-andre.com/ | email=
| address= | lat=48.87543 | long=2.31055 | directions={{station|Miromesnil|9|13}}
| phone= | tollfree= | fax=
| hours= | price=
| wikipedia=Musée Jacquemart-André | image=Musée Jacquemart André 2007 - Recoura.jpg | wikidata=Q1165526
| content=Private collection of French, Italian, Dutch masterpieces in a typical XIXth century mansion.
}}`
	srv := newServer(t, "Paris/8e", wikitext)
	defer srv.Close()

	p := wikivoyage.NewWithURL(srv.URL)
	pois, err := p.Search(context.Background(), types.SearchQuery{Mode: types.ModeDistrict, District: "Paris/8e"})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	var wv, wp *types.RawPoi
	for i := range pois {
		switch pois[i].Provider {
		case types.ProviderWikivoyage:
			wv = &pois[i]
		case types.ProviderWikipedia:
			wp = &pois[i]
		}
	}
	if wv == nil {
		t.Fatal("missing wikivoyage POI")
	}
	if wp == nil {
		t.Fatal("missing synthetic wikipedia POI — the nested template must not truncate the listing")
	}
	wantDesc := "Private collection of French, Italian, Dutch masterpieces in a typical XIXth century mansion."
	if wv.Description != wantDesc {
		t.Errorf("wikivoyage Description = %q, want %q", wv.Description, wantDesc)
	}
	if len(wv.Images) != 1 {
		t.Errorf("wikivoyage Images len = %d, want 1", len(wv.Images))
	}
	if !strings.Contains(wp.SourceURL, "Mus%C3%A9e_Jacquemart-Andr%C3%A9") {
		t.Errorf("wikipedia SourceURL = %q, want it to point at the article", wp.SourceURL)
	}
}

func TestName(t *testing.T) {
	p := wikivoyage.New("en")
	if p.Name() != types.ProviderWikivoyage {
		t.Errorf("Name() = %q, want %q", p.Name(), types.ProviderWikivoyage)
	}
}

// ── stripDescriptionMarkup (tested via Search) ───────────────────────────────

func TestSearch_DescriptionMarkupStripping(t *testing.T) {
	cases := []struct {
		desc    string
		content string
		want    string
	}{
		{
			desc:    "ref tag with inner content removed",
			content: `Iconic tower<ref>Bingham 2010</ref>.`,
			want:    "Iconic tower.",
		},
		{
			desc:    "self-closing ref removed",
			content: `Iconic tower<ref name="x"/>.`,
			want:    "Iconic tower.",
		},
		{
			desc:    "wiki link without pipe keeps target",
			content: `See the [[museum]].`,
			want:    "See the museum.",
		},
		{
			desc:    "bare html tag becomes space",
			content: `Line one<br/>line two.`,
			want:    "Line one line two.",
		},
		{
			desc:    "bold and italic markup removed",
			content: `'''Important''' and ''subtle'' detail.`,
			want:    "Important and subtle detail.",
		},
		{
			desc:    "multiple whitespace collapsed",
			content: `Iconic    tower.`,
			want:    "Iconic tower.",
		},
		{
			desc:    "html entities decoded",
			content: `Musée du quai Branly &mdash; Jacques Chirac.`,
			want:    "Musée du quai Branly — Jacques Chirac.",
		},
		{
			desc:    "realistic quai branly leakage",
			content: `Musée du quai Branly &mdash; Jacques Chirac<ref name="b">b</ref>. Designed by [[Jean Nouvel]]. ''Open daily.''<br/>Entry: 12&euro;.`,
			want:    "Musée du quai Branly — Jacques Chirac. Designed by Jean Nouvel. Open daily. Entry: 12€.",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			wikitext := `{{see|name=X|lat=48|long=2|content=` + tc.content + `}}`
			srv := newServer(t, "Zone", wikitext)
			defer srv.Close()

			p := wikivoyage.NewWithURL(srv.URL)
			pois, err := p.Search(context.Background(), types.SearchQuery{
				Mode: types.ModeDistrict, District: "Zone",
			})
			if err != nil {
				t.Fatalf("Search: %v", err)
			}
			if len(pois) != 1 {
				t.Fatalf("len = %d, want 1", len(pois))
			}
			if pois[0].Description != tc.want {
				t.Errorf("Description = %q, want %q", pois[0].Description, tc.want)
			}
		})
	}
}

// ── stripWikiMarkup (tested via Search) ──────────────────────────────────────

func TestSearch_WikiMarkupStripping(t *testing.T) {
	cases := []struct {
		desc     string
		wikitext string
		wantName string
		wantLen  int
	}{
		{
			// fieldRe captures up to `|`, so [[Article|Display]] is truncated to [[Article —
			// a broken fragment that stripWikiMarkup clears → empty name → POI dropped.
			desc:     "piped link truncated by field delimiter is dropped",
			wikitext: `{{see|name=[[Article|Musée d'Orsay]]|lat=48.86|long=2.32}}`,
			wantLen:  0,
		},
		{
			desc:     "plain link without display",
			wikitext: `{{see|name=[[Notre-Dame de Paris]]|lat=48.85|long=2.35}}`,
			wantName: "Notre-Dame de Paris",
			wantLen:  1,
		},
		{
			desc:     "anchor link strips anchor",
			wikitext: `{{see|name=[[Paris#History]]|lat=48.85|long=2.35}}`,
			wantName: "Paris",
			wantLen:  1,
		},
		{
			desc:     "namespace link keeps last segment",
			wikitext: `{{see|name=[[Paris/Marais]]|lat=48.85|long=2.35}}`,
			wantName: "Marais",
			wantLen:  1,
		},
		{
			desc:     "broken fragment truncated by pipe is dropped",
			wikitext: `{{see|name=[[Paris/4th|lat=48.85|long=2.35}}`,
			wantLen:  0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			srv := newServer(t, "Zone", tc.wikitext)
			defer srv.Close()

			p := wikivoyage.NewWithURL(srv.URL)
			pois, err := p.Search(context.Background(), types.SearchQuery{
				Mode: types.ModeDistrict, District: "Zone",
			})
			if err != nil {
				t.Fatalf("Search: %v", err)
			}
			if len(pois) != tc.wantLen {
				t.Fatalf("len = %d, want %d", len(pois), tc.wantLen)
			}
			if tc.wantLen > 0 && pois[0].Name != tc.wantName {
				t.Errorf("name = %q, want %q", pois[0].Name, tc.wantName)
			}
		})
	}
}

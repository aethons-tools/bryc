package robot

import (
	"errors"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func mustParse(t *testing.T, raw string) Request {
	t.Helper()
	q, err := url.ParseQuery(raw)
	if err != nil {
		t.Fatal(err)
	}
	req, err := ParseQuery(q)
	if err != nil {
		t.Fatalf("ParseQuery(%q): %v", raw, err)
	}
	return req
}

func TestParseQueryEmpty(t *testing.T) {
	req := mustParse(t, "")
	if !reflect.DeepEqual(req.Spec, Spec{}) {
		t.Errorf("Spec = %+v, want zero", req.Spec)
	}
	if req.Seed != nil || req.Palette != "" || req.Size != DefaultSize {
		t.Errorf("got seed=%v palette=%q size=%d", req.Seed, req.Palette, req.Size)
	}
}

func TestParseQueryRandomMeansUnset(t *testing.T) {
	req := mustParse(t, "head=random&rivets=random&body=&palette=")
	if !reflect.DeepEqual(req.Spec, Spec{}) || req.Palette != "" {
		t.Errorf("got %+v palette=%q, want all unset", req.Spec, req.Palette)
	}
}

func TestParseQueryFull(t *testing.T) {
	req := mustParse(t, "head=dome&eyes=visor&mouth=grille&antenna=bolt&ears=dials"+
		"&rivets=true&panels=false&blush=true"+
		"&body=%23AABBCC&accent=%23112233&glow=%23445566&background=none"+
		"&palette=mint&seed=42&size=256")
	s := req.Spec
	if s.Head != "dome" || s.Eyes != "visor" || s.Mouth != "grille" || s.Antenna != "bolt" || s.Ears != "dials" {
		t.Errorf("enums wrong: %+v", s)
	}
	if !*s.Rivets || *s.Panels || !*s.Blush {
		t.Errorf("bools wrong: rivets=%v panels=%v blush=%v", *s.Rivets, *s.Panels, *s.Blush)
	}
	if s.Body != "#aabbcc" || s.Accent != "#112233" || s.Glow != "#445566" || s.Background != "none" {
		t.Errorf("colors wrong (want lowercase): %+v", s)
	}
	if req.Palette != "mint" || req.Seed == nil || *req.Seed != 42 || req.Size != 256 {
		t.Errorf("got palette=%q seed=%v size=%d", req.Palette, req.Seed, req.Size)
	}
}

func TestParseQueryAggregatesProblems(t *testing.T) {
	q, _ := url.ParseQuery("head=blob&body=red&rivets=maybe&size=10&seed=-1&palette=nope&wat=1")
	_, err := ParseQuery(q)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("err = %v, want *ValidationError", err)
	}
	if len(ve.Problems) != 7 {
		t.Errorf("got %d problems, want 7:\n%v", len(ve.Problems), err)
	}
	for _, want := range []string{
		`head: unknown value "blob" (allowed: square, rounded, dome, trapezoid)`,
		`body: "red" is not a #rrggbb color`,
		`rivets: "maybe" is not true or false`,
		`size: 10 is outside 64-2048`,
		`seed: "-1" is not a non-negative integer`,
		`palette: unknown value "nope"`,
		`wat: unknown option`,
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error missing %q:\n%v", want, err)
		}
	}
}

func TestQueryRoundTrip(t *testing.T) {
	req := mustParse(t, "head=square&rivets=false&body=%23010203&background=none")
	again, err := ParseQuery(req.Spec.Query())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(again.Spec, req.Spec) {
		t.Errorf("round trip: got %+v, want %+v", again.Spec, req.Spec)
	}
	if got := req.Spec.Query().Encode(); got != "background=none&body=%23010203&head=square&rivets=false" {
		t.Errorf("Query().Encode() = %q", got)
	}
}

func TestFacets(t *testing.T) {
	fs := Facets()
	var names []string
	for _, f := range fs {
		names = append(names, f.Name)
	}
	want := "head eyes mouth antenna ears rivets panels blush body accent glow background"
	if strings.Join(names, " ") != want {
		t.Errorf("facet order = %v", names)
	}
	if fs[0].Kind != KindEnum || !reflect.DeepEqual(fs[0].Values, HeadValues) {
		t.Errorf("head facet = %+v", fs[0])
	}
	if fs[5].Kind != KindBool || fs[11].Kind != KindColor {
		t.Errorf("kinds wrong: %+v %+v", fs[5], fs[11])
	}
}

func TestPaletteNamesSorted(t *testing.T) {
	got := strings.Join(PaletteNames(), ",")
	if got != "bubblegum,chrome,midnight,mint,rusty,sunny" {
		t.Errorf("PaletteNames() = %s", got)
	}
}

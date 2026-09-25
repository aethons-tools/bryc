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
	req := mustParse(t, "head=dome&eyes=visor&mouth=jaw&expression=frown&antenna=bolt&ears=dials"+
		"&rivets=true&panels=false&blush=true"+
		"&body=%23AABBCC&accent=%23112233&glow=%23445566&background=none"+
		"&palette=mint&seed=42&size=256")
	s := req.Spec
	if s.Head != "dome" || s.Eyes != "visor" || s.Mouth != "jaw" || s.Expression != "frown" || s.Antenna != "bolt" || s.Ears != "dials" {
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
	req := mustParse(t, "head=square&rivets=false&body=%23010203&background=none&shoulders=0")
	again, err := ParseQuery(req.Spec.Query())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(again.Spec, req.Spec) {
		t.Errorf("round trip: got %+v, want %+v", again.Spec, req.Spec)
	}
	if got := req.Spec.Query().Encode(); got != "background=none&body=%23010203&head=square&rivets=false&shoulders=0" {
		t.Errorf("Query().Encode() = %q", got)
	}
}

func TestFacets(t *testing.T) {
	fs := Facets()
	byName := map[string]Facet{}
	var names []string
	for _, f := range fs {
		names = append(names, f.Name)
		byName[f.Name] = f
	}
	want := "head eyes eyecount mouth expression face antenna ears shoulders rivets panels blush body accent glow background"
	if strings.Join(names, " ") != want {
		t.Errorf("facet order = %v", names)
	}
	enums := map[string][]string{
		"head":       HeadValues,
		"mouth":      {"grille", "slot", "line", "jaw"},
		"expression": {"flat", "smile", "frown"},
		"eyecount":   {"2", "4", "6"},
	}
	for name, values := range enums {
		if f := byName[name]; f.Kind != KindEnum || !reflect.DeepEqual(f.Values, values) {
			t.Errorf("%s facet = %+v", name, f)
		}
	}
	ranges := map[string][2]int{"shoulders": {ShouldersMin, ShouldersMax}, "face": {FaceMin, FaceMax}}
	for name, r := range ranges {
		if f := byName[name]; f.Kind != KindRange || *f.Min != r[0] || *f.Max != r[1] {
			t.Errorf("%s facet = %+v", name, f)
		}
	}
	if d := byName["eyecount"].DependsOn; d == nil || *d != (Dependency{Facet: "eyes", Value: "round"}) {
		t.Errorf("eyecount dependsOn = %+v, want eyes=round", d)
	}
	if byName["head"].DependsOn != nil {
		t.Errorf("head should not depend on anything")
	}
	if byName["rivets"].Kind != KindBool || byName["background"].Kind != KindColor {
		t.Errorf("kinds wrong: %+v %+v", byName["rivets"], byName["background"])
	}
}

func TestPaletteNamesSorted(t *testing.T) {
	got := strings.Join(PaletteNames(), ",")
	if got != "bubblegum,chrome,midnight,mint,rusty,sunny" {
		t.Errorf("PaletteNames() = %s", got)
	}
}

func TestParseQueryShoulders(t *testing.T) {
	for raw, want := range map[string]int{"shoulders=-40": -40, "shoulders=0": 0, "shoulders=100": 100} {
		req := mustParse(t, raw)
		if req.Spec.Shoulders == nil || *req.Spec.Shoulders != want {
			t.Errorf("%s: Shoulders = %v, want %d", raw, req.Spec.Shoulders, want)
		}
	}
	for raw, want := range map[string]string{
		"shoulders=150":  "shoulders: 150 is outside -100 to 100",
		"shoulders=-101": "shoulders: -101 is outside -100 to 100",
		"shoulders=wide": `shoulders: "wide" is not an integer`,
		"face=3":         "face: 3 is outside -1 to 2",
		"face=-2":        "face: -2 is outside -1 to 2",
	} {
		q, _ := url.ParseQuery(raw)
		if _, err := ParseQuery(q); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%s: err = %v, want %q", raw, err, want)
		}
	}
}

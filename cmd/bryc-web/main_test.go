package main

import (
	"bytes"
	"encoding/json"
	"image/png"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/aethons-tools/bryc/robot"
)

func get(t *testing.T, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	newHandler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestRobotSeeded(t *testing.T) {
	rec := get(t, "/robot.png?seed=3&size=64&head=dome")
	if rec.Code != 200 || rec.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("code=%d type=%q body=%s", rec.Code, rec.Header().Get("Content-Type"), rec.Body)
	}
	if rec.Header().Get("X-Bryc-Seed") != "3" {
		t.Errorf("seed header = %q", rec.Header().Get("X-Bryc-Seed"))
	}
	if rec.Header().Get("Cache-Control") == "no-store" {
		t.Error("seeded response should be cacheable")
	}
	img, err := png.Decode(rec.Body)
	if err != nil || img.Bounds().Dx() != 64 {
		t.Fatalf("decode: %v", err)
	}
	q, err := url.ParseQuery(rec.Header().Get("X-Bryc-Spec"))
	if err != nil || q.Get("head") != "dome" || len(q) != len(robot.Facets()) {
		t.Errorf("X-Bryc-Spec = %q", rec.Header().Get("X-Bryc-Spec"))
	}
}

func TestRobotUnseeded(t *testing.T) {
	rec := get(t, "/robot.png?size=64")
	if rec.Code != 200 || rec.Header().Get("Cache-Control") != "no-store" || rec.Header().Get("X-Bryc-Seed") == "" {
		t.Errorf("code=%d headers=%v", rec.Code, rec.Header())
	}
}

func TestRobotInvalid(t *testing.T) {
	rec := get(t, "/robot.png?head=blob")
	if rec.Code != 400 || !strings.Contains(rec.Body.String(), `head: unknown value "blob"`) {
		t.Errorf("code=%d body=%s", rec.Code, rec.Body)
	}
}

func TestOptions(t *testing.T) {
	rec := get(t, "/options.json")
	var got struct {
		Facets   []robot.Facet `json:"facets"`
		Palettes []string      `json:"palettes"`
		Size     struct{ Default, Min, Max int }
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Facets) != len(robot.Facets()) || len(got.Palettes) == 0 || got.Size.Default != 512 {
		t.Errorf("got %+v", got)
	}
}

func TestIndex(t *testing.T) {
	rec := get(t, "/")
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "Bored Robots Yacht Club") {
		t.Errorf("code=%d body=%s", rec.Code, rec.Body)
	}
	for _, want := range []string{"/options.json", "/robot.png?", "X-Bryc-Seed", "X-Bryc-Spec", `id="size"`, "f.kind === 'range'", "dependsOn", "Randomize", `id="grid"`, `id="mode-grid"`, `id="mode-single"`, "'cell'", "randomSeed"} {
		if !strings.Contains(rec.Body.String(), want) {
			t.Errorf("index.html missing %q", want)
		}
	}
}

func TestRobotCell(t *testing.T) {
	a := get(t, "/robot.png?seed=5&cell=3&size=64")
	b := get(t, "/robot.png?seed=5&cell=3&size=64")
	if a.Code != 200 {
		t.Fatalf("code=%d body=%s", a.Code, a.Body)
	}
	want := strconv.FormatUint(robot.CellSeed(5, 3), 10)
	if got := a.Header().Get("X-Bryc-Seed"); got != want {
		t.Errorf("X-Bryc-Seed = %s, want the cell's own seed %s", got, want)
	}
	if !bytes.Equal(a.Body.Bytes(), b.Body.Bytes()) {
		t.Error("same grid seed and cell gave different robots")
	}
}

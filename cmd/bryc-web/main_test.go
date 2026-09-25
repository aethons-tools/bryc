package main

import (
	"encoding/json"
	"image/png"
	"net/http"
	"net/http/httptest"
	"net/url"
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
	if err != nil || q.Get("head") != "dome" || len(q) != 12 {
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
	if len(got.Facets) != 12 || len(got.Palettes) == 0 || got.Size.Default != 512 {
		t.Errorf("got %+v", got)
	}
}

func TestIndex(t *testing.T) {
	rec := get(t, "/")
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "Bored Robots Yacht Club") {
		t.Errorf("code=%d body=%s", rec.Code, rec.Body)
	}
	for _, want := range []string{"/options.json", "/robot.png?", "X-Bryc-Seed", "X-Bryc-Spec", `id="size"`, "Randomize"} {
		if !strings.Contains(rec.Body.String(), want) {
			t.Errorf("index.html missing %q", want)
		}
	}
}

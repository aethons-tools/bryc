// Command bryc-web serves the Bored Robots Yacht Club robot designer.
package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"flag"
	"image/png"
	"io/fs"
	"log"
	"net/http"
	"strconv"

	"github.com/aethons-tools/bryc/robot"
)

//go:embed static
var static embed.FS

func main() {
	addr := flag.String("addr", "localhost:8080", "listen address")
	flag.Parse()
	log.Printf("Bored Robots Yacht Club: http://%s", *addr)
	log.Fatal(http.ListenAndServe(*addr, newHandler()))
}

func newHandler() http.Handler {
	page, err := fs.Sub(static, "static")
	if err != nil {
		panic(err) // the embedded directory is fixed at build time
	}
	mux := http.NewServeMux()
	mux.Handle("GET /", http.FileServerFS(page))
	mux.HandleFunc("GET /robot.png", handleRobot)
	mux.HandleFunc("GET /robot.glb", handleModel)
	mux.HandleFunc("GET /options.json", handleOptions)
	return mux
}

// resolveRequest parses and resolves the query shared by /robot.png and
// /robot.glb and sets the seed, spec and cache headers. On invalid input it
// writes a 400 and reports false.
func resolveRequest(w http.ResponseWriter, r *http.Request) (robot.Spec, robot.Request, bool) {
	req, err := robot.ParseQuery(r.URL.Query())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return robot.Spec{}, req, false
	}
	spec, seed := req.Resolve()
	h := w.Header()
	h.Set("X-Bryc-Seed", strconv.FormatUint(seed, 10))
	h.Set("X-Bryc-Spec", spec.Query().Encode())
	if req.Seed == nil {
		h.Set("Cache-Control", "no-store")
	}
	return spec, req, true
}

func handleRobot(w http.ResponseWriter, r *http.Request) {
	spec, req, ok := resolveRequest(w, r)
	if !ok {
		return
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, robot.Render(spec, req.Size)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Write(buf.Bytes())
}

func handleModel(w http.ResponseWriter, r *http.Request) {
	spec, _, ok := resolveRequest(w, r)
	if !ok {
		return
	}
	var buf bytes.Buffer
	if err := robot.WriteGLB(&buf, spec); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "model/gltf-binary")
	w.Write(buf.Bytes())
}

type sizeInfo struct {
	Default int `json:"default"`
	Min     int `json:"min"`
	Max     int `json:"max"`
}

func handleOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Facets   []robot.Facet `json:"facets"`
		Palettes []string      `json:"palettes"`
		Size     sizeInfo      `json:"size"`
	}{robot.Facets(), robot.PaletteNames(), sizeInfo{robot.DefaultSize, robot.MinSize, robot.MaxSize}})
}

package robot

import (
	"flag"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "rewrite golden images in testdata/")

func boolp(b bool) *bool { return &b }
func intp(n int) *int    { return &n }

func TestGolden(t *testing.T) {
	cases := map[string]Spec{
		"seed-1": Resolve(Spec{}, 1),
		"seed-2": Resolve(Spec{}, 2),
		"seed-3": Resolve(Spec{}, 3),
		"pinned-transparent": {
			Head: "dome", Tall: boolp(true), Eyes: "visor", EyeCount: "4", EyeSize: "2", EyeStyle: "glower", Mouth: "jaw", Expression: "smile", Antenna: "bolt", Ears: "dials",
			Face: intp(1), Shoulders: intp(-60), Rivets: boolp(true), Panels: boolp(true), Blush: boolp(true),
			Body: "#98e2c6", Accent: "#ff8fa3", Glow: "#3ae0ff", Background: "none",
		},
	}
	for name, spec := range cases {
		t.Run(name, func(t *testing.T) {
			got := Render(spec, 256)
			path := filepath.Join("testdata", name+".png")
			if *update {
				writePNG(t, path, got)
				return
			}
			want := readPNG(t, path)
			if got.Bounds() != want.Bounds() {
				t.Fatalf("bounds %v, want %v", got.Bounds(), want.Bounds())
			}
			for y := 0; y < 256; y++ {
				for x := 0; x < 256; x++ {
					g := color.NRGBAModel.Convert(got.At(x, y))
					w := color.NRGBAModel.Convert(want.At(x, y))
					if g != w {
						t.Fatalf("pixel (%d,%d) = %v, want %v; if the change is intended run: go test ./robot -update", x, y, g, w)
					}
				}
			}
		})
	}
}

func writePNG(t *testing.T, path string, img image.Image) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

func readPNG(t *testing.T, path string) image.Image {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("%v (generate with: go test ./robot -update)", err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	return img
}

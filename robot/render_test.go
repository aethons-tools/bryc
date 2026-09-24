package robot

import (
	"image"
	"image/color"
	"testing"
)

// probe is a virtual-canvas point inside every head shape that no other part
// covers; see the plan's geometry notes.
const probeX, probeY = 360, 650

func pixel(img image.Image, size int, vx, vy float64) color.NRGBA {
	k := float64(size) / virtual
	return color.NRGBAModel.Convert(img.At(int(vx*k), int(vy*k))).(color.NRGBA)
}

func near(a, b color.NRGBA) bool {
	d := func(x, y uint8) bool { return max(x, y)-min(x, y) <= 2 }
	return d(a.R, b.R) && d(a.G, b.G) && d(a.B, b.B) && d(a.A, b.A)
}

func TestRenderSize(t *testing.T) {
	img := Render(Resolve(Spec{}, 1), 128)
	if b := img.Bounds(); b.Dx() != 128 || b.Dy() != 128 {
		t.Errorf("bounds = %v", b)
	}
}

func TestRenderBackground(t *testing.T) {
	s := Resolve(Spec{Background: "#ff0000"}, 1)
	if got := pixel(Render(s, 100), 100, 5, 5); got != (color.NRGBA{255, 0, 0, 255}) {
		t.Errorf("corner = %v, want red", got)
	}
	s.Background = "none"
	if got := pixel(Render(s, 100), 100, 5, 5); got.A != 0 {
		t.Errorf("corner = %v, want transparent", got)
	}
}

func TestRenderHeadUsesBodyColor(t *testing.T) {
	for _, head := range HeadValues {
		s := Resolve(Spec{Head: head, Body: "#3366cc"}, 3)
		if got := pixel(Render(s, 200), 200, probeX, probeY); !near(got, color.NRGBA{0x33, 0x66, 0xcc, 0xff}) {
			t.Errorf("head %s: probe = %v, want body color", head, got)
		}
	}
}

// TestRenderEveryValue renders each value of each enum and bool facet; it
// fails (panics) if any value has no drawing code.
func TestRenderEveryValue(t *testing.T) {
	for _, facet := range Facets() {
		values := facet.Values
		if facet.Kind == KindBool {
			values = []string{"true", "false"}
		}
		for _, v := range values {
			var s Spec
			for _, f := range s.fields() {
				if f.name != facet.Name {
					continue
				}
				if f.kind == KindBool {
					b := v == "true"
					*f.flag = &b
				} else {
					*f.str = v
				}
			}
			Render(Resolve(s, 11), 64)
		}
	}
}
func TestRenderCyclopsUsesGlow(t *testing.T) {
	s := Resolve(Spec{Eyes: "cyclops", Glow: "#00ff00"}, 4)
	// A point on the lens ring: right of center, between pupil (r40) and rim (r95).
	got := pixel(Render(s, 200), 200, 565, 444)
	if !near(got, color.NRGBA{0, 255, 0, 255}) {
		t.Errorf("lens = %v, want glow color", got)
	}
}

package robot

import (
	"image"
	"image/color"
	"net/url"
	"strconv"
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
		// The jaw mouth covers the probe by design, so pin a mouth that doesn't.
		s := Resolve(Spec{Head: head, Mouth: "line", Body: "#3366cc"}, 3)
		if got := pixel(Render(s, 200), 200, probeX, probeY); !near(got, color.NRGBA{0x33, 0x66, 0xcc, 0xff}) {
			t.Errorf("head %s: probe = %v, want body color", head, got)
		}
	}
}

// TestRenderEveryValue renders each value of each enum and bool facet and the
// ends and middle of each range facet; it fails (panics) if any value has no
// drawing code.
func TestRenderEveryValue(t *testing.T) {
	for _, facet := range Facets() {
		values := facet.Values
		switch facet.Kind {
		case KindBool:
			values = []string{"true", "false"}
		case KindRange:
			values = []string{strconv.Itoa(*facet.Min), "0", strconv.Itoa(*facet.Max)}
		}
		for _, v := range values {
			req, err := ParseQuery(url.Values{facet.Name: {v}})
			if err != nil {
				t.Fatal(err)
			}
			Render(Resolve(req.Spec, 11), 64)
		}
	}
}

// Shoulder probes, in virtual units: just inside the top-left corner of the
// shoulders, and above the shoulder line where the left spike rises.
const (
	cornerX, cornerY = 140, 840
	spikeX, spikeY   = 140, 780
)

func TestRenderShoulders(t *testing.T) {
	body, bg := color.NRGBA{0x33, 0x66, 0xcc, 0xff}, color.NRGBA{0xff, 0, 0, 0xff}
	render := func(shoulders int) image.Image {
		return Render(Resolve(Spec{Body: "#3366cc", Background: "#ff0000", Shoulders: &shoulders}, 3), 200)
	}
	cases := []struct {
		shoulders     int
		corner, spike color.NRGBA
	}{
		{0, body, bg},      // square corners, no spike
		{100, bg, bg},      // corner rounded away
		{-100, body, body}, // square corner with a spike above it
	}
	for _, c := range cases {
		img := render(c.shoulders)
		if got := pixel(img, 200, cornerX, cornerY); !near(got, c.corner) {
			t.Errorf("shoulders=%d: corner = %v, want %v", c.shoulders, got, c.corner)
		}
		if got := pixel(img, 200, spikeX, spikeY); !near(got, c.spike) {
			t.Errorf("shoulders=%d: spike probe = %v, want %v", c.shoulders, got, c.spike)
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

func TestRenderJaw(t *testing.T) {
	no := false
	body := color.NRGBA{0x33, 0x66, 0xcc, 0xff}
	jaw := shade(body, jawShade)
	for _, head := range HeadValues {
		s := Resolve(Spec{Head: head, Mouth: "jaw", Expression: "flat", Body: "#3366cc",
			Rivets: &no, Panels: &no, Blush: &no}, 3)
		img := Render(s, 1000)
		for _, p := range []struct {
			name string
			x, y float64
			want color.NRGBA
		}{
			{"chin", 360, 680, jaw},
			{"cheek band", 280, 600, jaw},
			{"face inside the U", 330, 580, body},
		} {
			if got := pixel(img, 1000, p.x, p.y); !near(got, p.want) {
				t.Errorf("head %s: %s (%v,%v) = %v, want %v", head, p.name, p.x, p.y, got, p.want)
			}
		}
	}
}

func TestRenderExpressionBendsLine(t *testing.T) {
	no := false
	render := func(expr string) image.Image {
		return Render(Resolve(Spec{Mouth: "line", Expression: expr, Body: "#3366cc", Blush: &no}, 3), 1000)
	}
	// The line's left end sits above center for a smile and below it for a frown.
	endX, above, below := 415.0, mouthY(headBox)-18, mouthY(headBox)+18
	smile, frown := render("smile"), render("frown")
	if got := pixel(smile, 1000, endX, above); !near(got, ink) {
		t.Errorf("smile: end above center = %v, want ink", got)
	}
	if got := pixel(frown, 1000, endX, below); !near(got, ink) {
		t.Errorf("frown: end below center = %v, want ink", got)
	}
	if got := pixel(smile, 1000, endX, below); near(got, ink) {
		t.Errorf("smile: end below center is ink, want face")
	}
}

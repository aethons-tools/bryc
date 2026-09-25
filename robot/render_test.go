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

// neutral pins the face position for tests that probe fixed coordinates.
var neutral = 0

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
		s := Resolve(Spec{Head: head, Mouth: "line", Face: &neutral, Body: "#3366cc"}, 3)
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

// TestRenderCyclopsHAL checks the HAL-style cyclops: a black lens with the
// glow lit in its center, inside a metal ring.
func TestRenderCyclopsHAL(t *testing.T) {
	s := Resolve(Spec{Eyes: "cyclops", Face: &neutral, Glow: "#00ff00"}, 4)
	ey, _ := facePos(s, headBox)
	img := Render(s, 1000)
	for _, p := range []struct {
		name string
		dx   float64 // distance right of the eye's center
		want color.NRGBA
	}{
		{"glow core", 12, color.NRGBA{0, 255, 0, 255}},
		{"dark lens", 72, screen},
		{"metal ring", 86, metal},
	} {
		if got := pixel(img, 1000, headBox.CX()+p.dx, ey); !near(got, p.want) {
			t.Errorf("%s at +%v = %v, want %v", p.name, p.dx, got, p.want)
		}
	}
}

func TestRenderJaw(t *testing.T) {
	no := false
	body := color.NRGBA{0x33, 0x66, 0xcc, 0xff}
	jaw := shade(body, jawShade)
	for _, head := range HeadValues {
		s := Resolve(Spec{Head: head, Mouth: "jaw", Expression: "flat", Face: &neutral, Body: "#3366cc",
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
		return Render(Resolve(Spec{Mouth: "line", Expression: expr, Face: &neutral, Body: "#3366cc", Blush: &no}, 3), 1000)
	}
	// The line's left end sits above center for a smile and below it for a frown.
	_, my := facePos(Spec{Head: "square", Face: &neutral}, headBox)
	endX, above, below := 415.0, my-18, my+18
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

func TestRenderFaceMovesEyes(t *testing.T) {
	no := false
	glow := color.NRGBA{0, 255, 0, 255}
	core := glow // the cyclops' solid glow core, just off its hot-spot center
	for _, face := range []int{-1, 0, 1, 2} {
		for _, head := range HeadValues {
			s := Resolve(Spec{Head: head, Eyes: "cyclops", Face: &face, Glow: "#00ff00",
				Rivets: &no, Panels: &no, Blush: &no}, 5)
			ey, _ := facePos(s, headBox)
			if got := pixel(Render(s, 1000), 1000, 512, ey); !near(got, core) {
				t.Errorf("face=%d head=%s: glow core at y=%v = %v, want %v", face, head, ey, got, core)
			}
		}
	}
	for _, head := range HeadValues {
		one, _ := facePos(Spec{Head: head, Face: ptr(1)}, headBox)
		two, _ := facePos(Spec{Head: head, Face: ptr(2)}, headBox)
		mid, _ := facePos(Spec{Head: head, Face: ptr(0)}, headBox)
		down, _ := facePos(Spec{Head: head, Face: ptr(-1)}, headBox)
		if !(one < mid && mid < down) {
			t.Errorf("%s: eye heights +1/0/-1 = %v/%v/%v, want increasing downward", head, one, mid, down)
		}
		if one != two {
			t.Errorf("%s: eyes at +2 (%v) should match +1 (%v)", head, two, one)
		}
	}
}

func TestFacePosMouth(t *testing.T) {
	for _, head := range HeadValues {
		_, m0 := facePos(Spec{Head: head, Face: ptr(0)}, headBox)
		_, m1 := facePos(Spec{Head: head, Face: ptr(1)}, headBox)
		_, m2 := facePos(Spec{Head: head, Face: ptr(2)}, headBox)
		_, mDown := facePos(Spec{Head: head, Face: ptr(-1)}, headBox)
		if m1 != m0 {
			t.Errorf("%s: face=+1 moved the mouth: %v -> %v", head, m0, m1)
		}
		if m2 >= m1 {
			t.Errorf("%s: face=+2 mouth at %v, want above +1's %v", head, m2, m1)
		}
		if mDown <= m0 {
			t.Errorf("%s: face=-1 mouth at %v, want below %v", head, mDown, m0)
		}
	}
}

// TestRenderFaceTopNoCollision checks that at face=+2, where the mouth moves
// up toward the eyes, the tallest eye (cyclops) and tallest mouth (grille)
// still have face showing between them.
func TestRenderFaceTopNoCollision(t *testing.T) {
	no, top := false, 2
	body := color.NRGBA{0x33, 0x66, 0xcc, 0xff}
	for _, head := range HeadValues {
		for _, expr := range ExpressionValues {
			s := Resolve(Spec{Head: head, Eyes: "cyclops", Mouth: "grille", Expression: expr, Face: &top,
				Body: "#3366cc", Glow: "#3366cc", Rivets: &no, Panels: &no, Blush: &no}, 5)
			ey, my := facePos(s, headBox)
			eyeBottom := ey + cyclopsRadius + outline/2
			mouthTop := my - grilleHalfHeight - outline/2 // the grille's center never bends
			if mouthTop-eyeBottom < 8 {
				t.Errorf("%s/%s: only %v units between eye and mouth", head, expr, mouthTop-eyeBottom)
				continue
			}
			if got := pixel(Render(s, 1000), 1000, headBox.CX(), (eyeBottom+mouthTop)/2); !near(got, body) {
				t.Errorf("%s/%s: between eye and mouth = %v, want face", head, expr, got)
			}
		}
	}
}

// TestRenderFaceFitsHead checks that at the highest face position the widest
// eye style (the visor) and the tallest (the cyclops) stay inside every head
// shape: just outside each one's outline must still be head, not background
// or the head's own outline.
func TestRenderFaceFitsHead(t *testing.T) {
	no, top := false, 2
	body := color.NRGBA{0x33, 0x66, 0xcc, 0xff}
	const gap = outline/2 + 4 // just past the eye's outline
	for _, head := range HeadValues {
		for _, eyes := range []string{"visor", "cyclops"} {
			s := Resolve(Spec{Head: head, Eyes: eyes, Face: &top, Ears: "none", Antenna: "none",
				Body: "#3366cc", Glow: "#3366cc", Background: "#ff0000", Rivets: &no, Panels: &no, Blush: &no}, 5) // glow = body hides the halo
			ey, _ := facePos(s, headBox)
			cx := headBox.CX()
			probes := [][2]float64{{cx, ey - cyclopsRadius - gap}}
			if eyes == "visor" {
				probes = [][2]float64{{cx - visorHalfWidth - gap, ey}, {cx + visorHalfWidth + gap, ey}}
			}
			img := Render(s, 1000)
			for _, p := range probes {
				if got := pixel(img, 1000, p[0], p[1]); !near(got, body) {
					t.Errorf("head %s, %s eyes: beside eye (%v,%v) = %v, want head color", head, eyes, p[0], p[1], got)
				}
			}
		}
	}
}

func TestRenderJawFollowsFace(t *testing.T) {
	no := false
	body := color.NRGBA{0x33, 0x66, 0xcc, 0xff}
	jaw := shade(body, jawShade)
	render := func(face int) image.Image {
		return Render(Resolve(Spec{Head: "square", Mouth: "jaw", Expression: "flat", Face: &face,
			Body: "#3366cc", Rivets: &no, Panels: &no, Blush: &no}, 3), 1000)
	}
	// A point on the jaw's left arm at neutral is above the arm once squashed.
	if got := pixel(render(0), 1000, 280, 545); !near(got, jaw) {
		t.Errorf("face=0: arm = %v, want jaw", got)
	}
	if got := pixel(render(-1), 1000, 280, 545); !near(got, body) {
		t.Errorf("face=-1: above arm = %v, want face", got)
	}
}

func ptr(n int) *int { return &n }

// TestRenderRoundEyeGlow checks the round eyes: a pale hot spot in the center
// and the glow drawn over the outline, so the outline is tinted, not pure ink.
func TestRenderRoundEyeGlow(t *testing.T) {
	glow := color.NRGBA{0, 255, 0, 255}
	s := Resolve(Spec{Eyes: "round", EyeCount: "2", EyeSize: "3", Face: &neutral, Glow: "#00ff00"}, 4)
	ey, _ := facePos(s, headBox)
	img := Render(s, 1000)
	x := headBox.CX() + eyeGap // right eye; the glint sits up-left of center
	if got, want := pixel(img, 1000, x, ey), hotSpotColor(glow); !near(got, want) {
		t.Errorf("center = %v, want hot spot %v", got, want)
	}
	if got := pixel(img, 1000, x+roundEyeRadius, ey); near(got, ink) {
		t.Errorf("outline = %v, want tinted by the glow", got)
	}
}

// TestRenderRoundEyeCounts checks every round eye, for each count, has its
// hot spot where roundEyeCenters says it is.
func TestRenderRoundEyeCounts(t *testing.T) {
	glow := color.NRGBA{0, 255, 0, 255}
	for _, count := range EyeCountValues {
		s := Resolve(Spec{Eyes: "round", EyeCount: count, Face: &neutral, Glow: "#00ff00"}, 4)
		centers := roundEyeCenters(s, headBox)
		if want, _ := strconv.Atoi(count); len(centers) != want {
			t.Fatalf("count %s: %d centers", count, len(centers))
		}
		img := Render(s, 1000)
		for _, p := range centers {
			if got := pixel(img, 1000, p[0], p[1]); !near(got, hotSpotColor(glow)) {
				t.Errorf("count %s: eye at (%v,%v) = %v, want hot spot", count, p[0], p[1], got)
			}
		}
	}
}

// TestRoundEyesWithinFaceBounds checks every round eye layout fits inside the
// cyclops' height and the visor's width, the bounds the face-position fit and
// collision tests are built on.
func TestRoundEyesWithinFaceBounds(t *testing.T) {
	for _, count := range EyeCountValues {
		l := roundLayouts[count]
		height := float64(l.rows-1)*l.rowGap + 2*l.r
		if height > 2*cyclopsRadius {
			t.Errorf("count %s: %v tall, cyclops is %v", count, height, 2.0*cyclopsRadius)
		}
		if l.colGap+l.r > visorHalfWidth {
			t.Errorf("count %s: reaches %v from center, visor reaches %v", count, l.colGap+l.r, visorHalfWidth)
		}
	}
}

// TestRoundEyeRadius checks eye size picks the round-eye radius, capped at
// the largest size the eye count allows.
func TestRoundEyeRadius(t *testing.T) {
	cases := []struct {
		count, size string
		want        float64
	}{
		{"2", "3", 58}, {"2", "2", 40}, {"2", "1", 25},
		{"4", "3", 40}, {"4", "2", 40}, {"4", "1", 25},
		{"6", "3", 25}, {"6", "2", 25}, {"6", "1", 25},
	}
	for _, c := range cases {
		if got := roundEyeR(Spec{EyeCount: c.count, EyeSize: c.size}); got != c.want {
			t.Errorf("count %s size %s: radius %v, want %v", c.count, c.size, got, c.want)
		}
	}
}

func TestRenderEyeSizeShrinksEyes(t *testing.T) {
	no := false
	glow, body := color.NRGBA{0, 255, 0, 255}, color.NRGBA{0x33, 0x66, 0xcc, 0xff}
	render := func(size string) image.Image {
		return Render(Resolve(Spec{Eyes: "round", EyeCount: "2", EyeSize: size, Face: &neutral,
			Glow: "#00ff00", Body: "#3366cc", Blush: &no}, 4), 1000)
	}
	ey, _ := facePos(Spec{Head: "square", Face: &neutral}, headBox)
	x := headBox.CX() + eyeGap + 46 // inside a size-3 eye, clear of a size-1 eye and its glow
	if got := pixel(render("3"), 1000, x, ey); !near(got, glow) {
		t.Errorf("size 3: %v, want glow", got)
	}
	if got := pixel(render("1"), 1000, x, ey); !near(got, body) {
		t.Errorf("size 1: %v, want face", got)
	}
}

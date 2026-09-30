package robot

import (
	"fmt"
	"image/color"
	"math"
	"testing"

	"github.com/aethons-tools/bryc/robot/mesh"
)

// handExtent is the canvas bounding box of everything in h.
func handExtent(h hand) (lo, hi mesh.Vec2) {
	lo, hi = mesh.Vec2{X: math.Inf(1), Y: math.Inf(1)}, mesh.Vec2{X: math.Inf(-1), Y: math.Inf(-1)}
	grow := func(p mesh.Vec2, r float64) {
		q := h.at(p)
		lo = mesh.Vec2{X: math.Min(lo.X, q.X-r), Y: math.Min(lo.Y, q.Y-r)}
		hi = mesh.Vec2{X: math.Max(hi.X, q.X+r), Y: math.Max(hi.Y, q.Y+r)}
	}
	for _, p := range h.palm {
		grow(p, 0)
	}
	for _, d := range h.digits {
		grow(d.root, d.r)
		grow(d.tip, d.r)
	}
	return lo, hi
}

func TestHandsMirrored(t *testing.T) {
	for _, b := range []Layout{headBox, tallHeadBox} {
		hs := hands(b)
		l, r := hs[0], hs[1]
		if l.wrist.X+r.wrist.X != 2*b.CX() || l.wrist.Y != r.wrist.Y {
			t.Errorf("wrists %v %v not mirrored about %v", l.wrist, r.wrist, b.CX())
		}
		for i := range l.palm {
			if math.Abs(l.palm[i].X+r.palm[i].X) > 1e-9 || l.palm[i].Y != r.palm[i].Y {
				t.Fatalf("palm point %d: %v vs %v", i, l.palm[i], r.palm[i])
			}
		}
		// The thumb points toward the middle.
		if l.digits[2].tip.X <= l.digits[2].root.X || r.digits[2].tip.X >= r.digits[2].root.X {
			t.Errorf("thumbs point outward: %+v %+v", l.digits[2], r.digits[2])
		}
	}
}

func TestHandsClearOfMouthAndShadow(t *testing.T) {
	for _, head := range HeadValues {
		for _, tall := range []bool{false, true} {
			for face := FaceMin; face <= FaceMax; face++ {
				s := Resolve(Spec{Head: head, Tall: &tall, Face: &face}, 1)
				b := headLayout(s)
				mouthBottom := mouthY(s, b) + grilleHalfHeight + maxBend
				_, shadowY, _, shadowRY := floorShadow(b)
				for i, h := range hands(b) {
					lo, hi := handExtent(h)
					if lo.Y <= mouthBottom {
						t.Errorf("%s tall=%v face %d hand %d: top %.0f covers the mouth (bottom %.0f)", head, tall, face, i, lo.Y, mouthBottom)
					}
					if hi.Y+outline/2 >= shadowY-shadowRY {
						t.Errorf("%s tall=%v face %d hand %d: bottom %.0f touches the shadow", head, tall, face, i, hi.Y)
					}
				}
			}
		}
	}
}

func TestRenderHands(t *testing.T) {
	s := Resolve(Spec{Body: "#3366cc", Background: "#ff0000"}, 3)
	img := Render(s, 1000)
	body := color.NRGBA{0x33, 0x66, 0xcc, 0xff}
	for i, h := range hands(headLayout(s)) {
		// The middle of the palm's knuckle edge, and each fingertip.
		probes := []mesh.Vec2{h.at(mesh.Vec2{X: (h.digits[0].root.X + h.digits[1].root.X) / 2, Y: (h.digits[0].root.Y + h.digits[1].root.Y) / 2})}
		for _, d := range h.digits {
			probes = append(probes, h.at(d.tip))
		}
		for _, p := range probes {
			if got := pixel(img, 1000, p.X, p.Y); !near(got, body) {
				t.Errorf("hand %d at %v = %v, want body %v", i, p, got, body)
			}
		}
	}
}

func TestModelHands(t *testing.T) {
	for _, mouth := range MouthValues {
		s := Resolve(Spec{Mouth: mouth}, 4)
		nodes := nodeNames(s)
		for i := range 2 {
			for _, name := range []string{fmt.Sprintf("hand-%d", i), fmt.Sprintf("finger-%d-0", i), fmt.Sprintf("finger-%d-1", i), fmt.Sprintf("thumb-%d", i)} {
				m, ok := nodes[name]
				if !ok {
					t.Fatalf("%s: missing node %q", mouth, name)
				}
				if !m.Closed() || m.Volume() <= 0 {
					t.Errorf("%s %s: closed %v volume %v", mouth, name, m.Closed(), m.Volume())
				}
				// In front of the face, and of the jaw's front.
				if lo, _ := m.Bounds(); lo.Z <= jawFront*0.001 {
					t.Errorf("%s %s: back at z %.3f, behind the face", mouth, name, lo.Z)
				}
			}
		}
	}
}

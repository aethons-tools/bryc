package mesh

import (
	"math"
	"testing"
)

func circlePoly(r float64, n int) []Vec2 {
	var p []Vec2
	for i := 0; i < n; i++ {
		a := 2 * math.Pi * float64(i) / float64(n)
		p = append(p, Vec2{r * math.Cos(a), r * math.Sin(a)})
	}
	return p
}

func TestExtrudeClosedOutwardAndBounded(t *testing.T) {
	polys := map[string][]Vec2{
		"square":    {{0, 0}, {100, 0}, {100, 100}, {0, 100}},
		"clockwise": {{0, 0}, {0, 100}, {100, 100}, {100, 0}},
		"L":         {{0, 0}, {300, 0}, {300, 100}, {100, 100}, {100, 300}, {0, 300}},
		"circle":    circlePoly(80, 48),
		"dup-end":   {{0, 0}, {100, 0}, {100, 100}, {0, 100}, {0, 0}},
	}
	for name, poly := range polys {
		m, err := Extrude(poly, 60, 10)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if !m.Closed() {
			t.Errorf("%s: not closed", name)
		}
		if m.Volume() <= 0 {
			t.Errorf("%s: volume %v, want positive", name, m.Volume())
		}
		lo, hi := m.Bounds()
		var plo, phi Vec2 = Vec2{math.Inf(1), math.Inf(1)}, Vec2{math.Inf(-1), math.Inf(-1)}
		for _, p := range poly {
			plo = Vec2{math.Min(plo.X, p.X), math.Min(plo.Y, p.Y)}
			phi = Vec2{math.Max(phi.X, p.X), math.Max(phi.Y, p.Y)}
		}
		const e = 1e-9
		if lo.X < plo.X-e || lo.Y < plo.Y-e || hi.X > phi.X+e || hi.Y > phi.Y+e || math.Abs(lo.Z+30) > e || math.Abs(hi.Z-30) > e {
			t.Errorf("%s: bounds %v..%v outside outline %v..%v, z ±30", name, lo, hi, plo, phi)
		}
		for i, n := range m.Normals {
			if math.Abs(n.Len()-1) > 1e-9 {
				t.Fatalf("%s: normal %d not unit: %v", name, i, n)
			}
		}
	}
}

func TestExtrudeSquareVolume(t *testing.T) {
	// A 100×100×60 box with 10-radius rounded edges loses a little volume
	// along its 8 long edges; it must be less than the box but close.
	m, err := Extrude([]Vec2{{0, 0}, {100, 0}, {100, 100}, {0, 100}}, 60, 10)
	if err != nil {
		t.Fatal(err)
	}
	if v := m.Volume(); v >= 600000 || v < 580000 {
		t.Errorf("volume %v, want a bit under 600000", v)
	}
}

func TestExtrudeCollinearVertices(t *testing.T) {
	// A square with its edge midpoints included: 4 collinear vertices.
	poly := []Vec2{{0, 0}, {50, 0}, {100, 0}, {100, 50}, {100, 100}, {50, 100}, {0, 100}, {0, 50}}
	m, err := Extrude(poly, 60, 10)
	if err != nil {
		t.Fatal(err)
	}
	if !m.Closed() {
		t.Error("not closed")
	}
	if m.Volume() <= 0 {
		t.Errorf("volume %v, want positive", m.Volume())
	}
}

func TestExtrudeErrors(t *testing.T) {
	sq := []Vec2{{0, 0}, {10, 0}, {10, 10}, {0, 10}}
	for name, c := range map[string]struct {
		poly         []Vec2
		depth, bevel float64
	}{
		"no bevel":      {sq, 10, 0},
		"bevel too big": {sq, 10, 5},
		"two points":    {sq[:2], 10, 1},
	} {
		if _, err := Extrude(c.poly, c.depth, c.bevel); err == nil {
			t.Errorf("%s: want error", name)
		}
	}
}

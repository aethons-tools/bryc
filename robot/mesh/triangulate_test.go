package mesh

import (
	"math"
	"testing"
)

func triArea(poly []Vec2, idx []int) float64 {
	sum := 0.0
	for i := 0; i < len(idx); i += 3 {
		a, b, c := poly[idx[i]], poly[idx[i+1]], poly[idx[i+2]]
		area := ((b.X-a.X)*(c.Y-a.Y) - (c.X-a.X)*(b.Y-a.Y)) / 2
		if area < -1e-12 {
			return math.NaN() // a clockwise triangle
		}
		sum += area
	}
	return sum
}

func TestTriangulate(t *testing.T) {
	square := []Vec2{{0, 0}, {2, 0}, {2, 2}, {0, 2}}
	lshape := []Vec2{{0, 0}, {3, 0}, {3, 1}, {1, 1}, {1, 3}, {0, 3}}
	withCollinear := []Vec2{{0, 0}, {1, 0}, {2, 0}, {2, 2}, {0, 2}}
	cw := []Vec2{{0, 0}, {0, 2}, {2, 2}, {2, 0}}
	var circle []Vec2
	for i := 0; i < 64; i++ {
		a := 2 * math.Pi * float64(i) / 64
		circle = append(circle, Vec2{math.Cos(a), math.Sin(a)})
	}
	for name, poly := range map[string][]Vec2{"square": square, "L": lshape, "collinear": withCollinear, "clockwise": cw, "circle": circle} {
		idx, err := Triangulate(poly)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if got, want := triArea(poly, idx), math.Abs(Area(poly)); math.IsNaN(got) || math.Abs(got-want) > 1e-9 {
			t.Errorf("%s: triangles cover %v, polygon area %v", name, got, want)
		}
	}
}

func TestTriangulateRejectsDegenerate(t *testing.T) {
	if _, err := Triangulate([]Vec2{{0, 0}, {1, 0}}); err == nil {
		t.Error("two points should fail")
	}
	if _, err := Triangulate([]Vec2{{0, 0}, {1, 0}, {2, 0}}); err == nil {
		t.Error("zero-area polygon should fail")
	}
}

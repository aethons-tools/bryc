package robot

import (
	"fmt"
	"math"
	"testing"

	"github.com/aethons-tools/bryc/robot/mesh"
)

// crossings is where a horizontal line at y crosses the polygon's edges.
func crossings(poly []mesh.Vec2, y float64) []float64 {
	var xs []float64
	for i := range poly {
		a, b := poly[i], poly[(i+1)%len(poly)]
		if (a.Y <= y) != (b.Y <= y) {
			xs = append(xs, a.X+(y-a.Y)*(b.X-a.X)/(b.Y-a.Y))
		}
	}
	return xs
}

func TestHeadOutlineMatchesEdges(t *testing.T) {
	for _, head := range HeadValues {
		for _, tall := range []bool{false, true} {
			s := Spec{Head: head, Tall: &tall}
			b := headLayout(s)
			poly := headOutline(s, b)
			if a := mesh.Area(poly); math.Abs(a) < 1000 {
				t.Fatalf("%s tall=%v: area %v", head, tall, a)
			}
			for _, f := range []float64{0.15, 0.3, 0.5, 0.7, 0.85} {
				y := b.Y + b.H*f
				xs := crossings(poly, y)
				if len(xs) != 2 {
					t.Fatalf("%s tall=%v y=%v: %d crossings", head, tall, y, len(xs))
				}
				left := math.Min(xs[0], xs[1])
				if want := leftEdge(head, b, y); math.Abs(left-want) > 1.5 {
					t.Errorf("%s tall=%v y=%.0f: outline left %v, leftEdge %v", head, tall, y, left, want)
				}
			}
		}
	}
}

func TestShoulderOutline(t *testing.T) {
	for _, sh := range []int{-100, -40, 0, 40, 100} {
		poly := shoulderOutline(sh)
		lo, hi := math.Inf(1), math.Inf(-1)
		for _, p := range poly {
			lo, hi = math.Min(lo, p.Y), math.Max(hi, p.Y)
		}
		if hi != virtual {
			t.Errorf("shoulders %d: bottom at %v, want cut at %v", sh, hi, virtual)
		}
		if _, err := mesh.Triangulate(poly); err != nil {
			t.Errorf("shoulders %d: %v", sh, err)
		}
		wantTop := shoulderTop
		if sh < 0 {
			wantTop = shoulderTop - maxSpikeHeight*float64(-sh)/ShouldersMax
		}
		if math.Abs(lo-wantTop) > 1e-9 {
			t.Errorf("shoulders %d: top %v, want %v", sh, lo, wantTop)
		}
	}
}

func TestJawOutlineIsSimpleU(t *testing.T) {
	no := false
	for _, head := range HeadValues {
		for _, expr := range ExpressionValues {
			for face := FaceMin; face <= FaceMax; face++ {
				s := Spec{Head: head, Tall: &no, Expression: expr, Face: &face}
				b := headLayout(s)
				poly := jawOutline(s, b)
				if _, err := mesh.Triangulate(poly); err != nil {
					t.Fatalf("%s %s face %d: %v", head, expr, face, err)
				}
				// Across the arms (just below the jaw top) the U has four crossings.
				if xs := crossings(poly, jawTop(s, b)+10); len(xs) != 4 {
					t.Errorf("%s %s face %d: %d crossings below the top, want 4", head, expr, face, len(xs))
				}
			}
		}
	}
}

func TestBentRectOutlineMatchesCurve(t *testing.T) {
	p := bentRectOutline("smile", 500, 600, 240, 70)
	if len(p) != 50 {
		t.Fatalf("%d points, want 50", len(p))
	}
	if p[0].X != 380 || math.Abs(p[0].Y-(curve("smile", 500, 600, 240, 380)-35)) > 1e-9 {
		t.Errorf("first point %v", p[0])
	}
}

// crossingEdges reports a pair of non-adjacent edges of poly that properly
// cross, which mesh.Triangulate does not always catch.
func crossingEdges(poly []mesh.Vec2) (int, int, bool) {
	side := func(p, q, r mesh.Vec2) float64 { return (q.X-p.X)*(r.Y-p.Y) - (q.Y-p.Y)*(r.X-p.X) }
	opposite := func(a, b float64) bool { return (a > 1e-9 && b < -1e-9) || (a < -1e-9 && b > 1e-9) }
	n := len(poly)
	for i := range n {
		a, b := poly[i], poly[(i+1)%n]
		for j := i + 2; j < n; j++ {
			if i == 0 && j == n-1 {
				continue // adjacent through the closing edge
			}
			c, d := poly[j], poly[(j+1)%n]
			if opposite(side(c, d, a), side(c, d, b)) && opposite(side(a, b, c), side(a, b, d)) {
				return i, j, true
			}
		}
	}
	return 0, 0, false
}

func TestOutlinesHaveNoCrossingEdges(t *testing.T) {
	for _, tall := range []bool{false, true} {
		for _, head := range HeadValues {
			for _, expr := range ExpressionValues {
				for face := FaceMin; face <= FaceMax; face++ {
					s := Spec{Head: head, Tall: &tall, Expression: expr, Face: &face}
					b := headLayout(s)
					for name, poly := range map[string][]mesh.Vec2{"head": headOutline(s, b), "jaw": jawOutline(s, b)} {
						if i, j, bad := crossingEdges(poly); bad {
							t.Errorf("%s: %s tall=%v %s face %d: edges %d and %d cross", name, head, tall, expr, face, i, j)
						}
					}
				}
			}
		}
	}
	for sh := ShouldersMin; sh <= ShouldersMax; sh++ {
		if i, j, bad := crossingEdges(shoulderOutline(sh)); bad {
			t.Errorf("shoulders %d: edges %d and %d cross", sh, i, j)
		}
	}
}

// extrudes checks that poly extrudes into a closed solid, as the 3D model does.
func extrudes(t *testing.T, name string, poly []mesh.Vec2, depth, bevel float64) {
	t.Helper()
	m, err := mesh.Extrude(poly, depth, bevel)
	if err != nil {
		t.Errorf("%s: %v", name, err)
		return
	}
	if !m.Closed() || m.Volume() <= 0 {
		t.Errorf("%s: closed %v, volume %v", name, m.Closed(), m.Volume())
	}
}

func TestShoulderOutlineExtrudes(t *testing.T) {
	for sh := ShouldersMin; sh <= ShouldersMax; sh++ {
		extrudes(t, fmt.Sprint("shoulders ", sh), shoulderOutline(sh), 360, 12)
	}
}

func TestHeadOutlineExtrudes(t *testing.T) {
	for _, tall := range []bool{false, true} {
		for _, head := range HeadValues {
			s := Spec{Head: head, Tall: &tall}
			extrudes(t, fmt.Sprint(head, " tall=", tall), headOutline(s, headLayout(s)), 300, 12)
		}
	}
}

func TestJawOutlineExtrudes(t *testing.T) {
	for _, tall := range []bool{false, true} {
		for _, head := range HeadValues {
			for _, expr := range ExpressionValues {
				for face := FaceMin; face <= FaceMax; face++ {
					s := Spec{Head: head, Tall: &tall, Expression: expr, Face: &face}
					name := fmt.Sprint(head, " tall=", tall, " ", expr, " face ", face)
					extrudes(t, name, jawOutline(s, headLayout(s)), 330, 8)
				}
			}
		}
	}
}

package mesh

import (
	"errors"
	"math"
)

// Area is the polygon's signed area: positive when counter-clockwise.
func Area(poly []Vec2) float64 {
	a := 0.0
	for i := range poly {
		p, q := poly[i], poly[(i+1)%len(poly)]
		a += p.X*q.Y - q.X*p.Y
	}
	return a / 2
}

func cross2(o, a, b Vec2) float64 { return (a.X-o.X)*(b.Y-o.Y) - (a.Y-o.Y)*(b.X-o.X) }

func inTriangle(p, a, b, c Vec2) bool {
	return cross2(a, b, p) >= 0 && cross2(b, c, p) >= 0 && cross2(c, a, p) >= 0
}

// Triangulate splits a simple polygon (either winding) into triangles by ear
// clipping. It returns indices into poly, three per triangle, each triangle
// counter-clockwise.
func Triangulate(poly []Vec2) ([]int, error) {
	n := len(poly)
	area := Area(poly)
	if n < 3 || math.Abs(area) < 1e-12 {
		return nil, errors.New("mesh: degenerate polygon")
	}
	idx := make([]int, n)
	for i := range idx {
		idx[i] = i
	}
	if area < 0 { // work counter-clockwise
		for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
			idx[i], idx[j] = idx[j], idx[i]
		}
	}
	scale := 0.0
	for _, p := range poly {
		scale = math.Max(scale, math.Max(math.Abs(p.X), math.Abs(p.Y)))
	}
	eps := 1e-12 * math.Max(1, scale*scale)
	var tris []int
	for guard := 0; len(idx) > 3; guard++ {
		if guard > 4*n*n {
			return nil, errors.New("mesh: polygon is not simple")
		}
		clipped := false
		for i := range idx {
			ia, ib, ic := idx[(i+len(idx)-1)%len(idx)], idx[i], idx[(i+1)%len(idx)]
			a, b, c := poly[ia], poly[ib], poly[ic]
			turn := cross2(a, b, c)
			if math.Abs(turn) <= eps { // collinear: drop the middle vertex
				idx = append(idx[:i], idx[i+1:]...)
				clipped = true
				break
			}
			if turn < 0 {
				continue // reflex
			}
			ear := true
			for _, j := range idx {
				if j == ia || j == ib || j == ic {
					continue
				}
				if p := poly[j]; (p != a && p != b && p != c) && inTriangle(p, a, b, c) {
					ear = false
					break
				}
			}
			if ear {
				tris = append(tris, ia, ib, ic)
				idx = append(idx[:i], idx[i+1:]...)
				clipped = true
				break
			}
		}
		if !clipped {
			return nil, errors.New("mesh: polygon is not simple")
		}
	}
	if len(idx) == 3 && math.Abs(cross2(poly[idx[0]], poly[idx[1]], poly[idx[2]])) > eps {
		tris = append(tris, idx[0], idx[1], idx[2])
	}
	return tris, nil
}

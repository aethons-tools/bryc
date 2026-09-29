package mesh

import (
	"errors"
	"math"
)

// bevelSteps is how many segments round each edge's quarter circle.
const bevelSteps = 4

// Extrude pushes a 2D outline out into a closed slab depth thick, centred on
// z=0, with its side edges rounded by a quarter circle of radius bevel.
func Extrude(poly []Vec2, depth, bevel float64) (Mesh, error) {
	p := cleanPoly(poly)
	if len(p) < 3 || math.Abs(Area(p)) < 1e-12 {
		return Mesh{}, errors.New("mesh: degenerate outline")
	}
	if bevel <= 0 || bevel >= depth/2 {
		return Mesh{}, errors.New("mesh: bevel must be > 0 and < depth/2")
	}
	if Area(p) < 0 {
		for i, j := 0, len(p)-1; i < j; i, j = i+1, j-1 {
			p[i], p[j] = p[j], p[i]
		}
	}
	n := len(p)
	miter := make([]Vec2, n) // bisector scaled by the miter factor
	dir := make([]Vec2, n)   // unit bisector
	for i := range p {
		n1 := outward(p[(i+n-1)%n], p[i])
		n2 := outward(p[i], p[(i+1)%n])
		m := Vec2{n1.X + n2.X, n1.Y + n2.Y}
		l := math.Hypot(m.X, m.Y)
		if l < 1e-12 {
			m = n1
		} else {
			m = Vec2{m.X / l, m.Y / l}
		}
		f := 1 / math.Max(m.X*n1.X+m.Y*n1.Y, 0.5)
		dir[i], miter[i] = m, Vec2{m.X * f, m.Y * f}
	}
	var mesh Mesh
	ring := func(theta, side float64) uint32 {
		base := uint32(len(mesh.Positions))
		in := bevel * (1 - math.Cos(theta))
		z := side * (depth/2 - bevel + bevel*math.Sin(theta))
		for i := range p {
			mesh.Positions = append(mesh.Positions, Vec3{p[i].X - miter[i].X*in, p[i].Y - miter[i].Y*in, z})
			mesh.Normals = append(mesh.Normals, Vec3{dir[i].X * math.Cos(theta), dir[i].Y * math.Cos(theta), side * math.Sin(theta)}.Norm())
		}
		return base
	}
	var rings []uint32
	for k := bevelSteps; k >= 0; k-- {
		rings = append(rings, ring(math.Pi/2*float64(k)/bevelSteps, -1))
	}
	for k := 0; k <= bevelSteps; k++ {
		rings = append(rings, ring(math.Pi/2*float64(k)/bevelSteps, +1))
	}
	for r := 0; r+1 < len(rings); r++ {
		a, b := rings[r], rings[r+1]
		for i := 0; i < n; i++ {
			j := (i + 1) % n
			mesh.Indices = append(mesh.Indices, a+uint32(i), a+uint32(j), b+uint32(j), a+uint32(i), b+uint32(j), b+uint32(i))
		}
	}
	// Caps: the fully inset outline.
	inset := make([]Vec2, n)
	for i := range p {
		inset[i] = Vec2{p[i].X - miter[i].X*bevel, p[i].Y - miter[i].Y*bevel}
	}
	insetArea := Area(inset)
	if insetArea <= 1e-9*math.Abs(Area(p)) {
		return Mesh{}, errors.New("mesh: outline too thin for bevel")
	}
	tris, err := Triangulate(inset)
	if err != nil {
		return Mesh{}, err
	}
	used := make([]bool, n)
	triArea := 0.0
	for t := 0; t+2 < len(tris); t += 3 {
		used[tris[t]], used[tris[t+1]], used[tris[t+2]] = true, true, true
		triArea += math.Abs(Area([]Vec2{inset[tris[t]], inset[tris[t+1]], inset[tris[t+2]]}))
	}
	for _, u := range used {
		if !u {
			return Mesh{}, errors.New("mesh: inset outline has an unusable vertex")
		}
	}
	if math.Abs(triArea-insetArea) > 1e-6*insetArea {
		return Mesh{}, errors.New("mesh: outline too thin for bevel")
	}
	for _, side := range []float64{1, -1} {
		base := uint32(len(mesh.Positions))
		for _, q := range inset {
			mesh.Positions = append(mesh.Positions, Vec3{q.X, q.Y, side * depth / 2})
			mesh.Normals = append(mesh.Normals, Vec3{0, 0, side})
		}
		for t := 0; t < len(tris); t += 3 {
			a, b, c := base+uint32(tris[t]), base+uint32(tris[t+1]), base+uint32(tris[t+2])
			if side < 0 {
				b, c = c, b
			}
			mesh.Indices = append(mesh.Indices, a, b, c)
		}
	}
	return mesh, nil
}

// outward is the unit outward normal of edge a→b of a counter-clockwise
// polygon.
func outward(a, b Vec2) Vec2 {
	dx, dy := b.X-a.X, b.Y-a.Y
	l := math.Hypot(dx, dy)
	return Vec2{dy / l, -dx / l}
}

// cleanPoly copies poly without consecutive duplicate points (including a
// closing point equal to the first) and without collinear vertices, which
// would otherwise leave T-junctions between the cap and the side wall.
func cleanPoly(poly []Vec2) []Vec2 {
	var out []Vec2
	for _, q := range poly {
		if len(out) > 0 && math.Hypot(q.X-out[len(out)-1].X, q.Y-out[len(out)-1].Y) < 1e-9 {
			continue
		}
		out = append(out, q)
	}
	for len(out) > 1 && math.Hypot(out[0].X-out[len(out)-1].X, out[0].Y-out[len(out)-1].Y) < 1e-9 {
		out = out[:len(out)-1]
	}
	for changed := true; changed && len(out) >= 3; {
		changed = false
		for i := 0; i < len(out); i++ {
			prev, v, next := out[(i+len(out)-1)%len(out)], out[i], out[(i+1)%len(out)]
			e1x, e1y := v.X-prev.X, v.Y-prev.Y
			e2x, e2y := next.X-v.X, next.Y-v.Y
			cr := e1x*e2y - e1y*e2x
			if math.Abs(cr) < 1e-9*math.Hypot(e1x, e1y)*math.Hypot(e2x, e2y) {
				out = append(out[:i], out[i+1:]...)
				changed = true
				break
			}
		}
	}
	return out
}

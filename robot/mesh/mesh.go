package mesh

import "math"

// Mesh is an indexed triangle mesh with per-vertex normals. Triangles are
// counter-clockwise seen from outside.
type Mesh struct {
	Positions []Vec3
	Normals   []Vec3
	Indices   []uint32
}

// Append adds o's triangles to m.
func (m *Mesh) Append(o Mesh) {
	base := uint32(len(m.Positions))
	m.Positions = append(m.Positions, o.Positions...)
	m.Normals = append(m.Normals, o.Normals...)
	for _, i := range o.Indices {
		m.Indices = append(m.Indices, base+i)
	}
}

// mapped returns a copy of m with f applied to every position and normal.
func (m Mesh) mapped(f func(p, n Vec3) (Vec3, Vec3)) Mesh {
	out := Mesh{
		Positions: make([]Vec3, len(m.Positions)),
		Normals:   make([]Vec3, len(m.Normals)),
		Indices:   append([]uint32(nil), m.Indices...),
	}
	for i := range m.Positions {
		out.Positions[i], out.Normals[i] = f(m.Positions[i], m.Normals[i])
	}
	return out
}

func (m Mesh) Translate(d Vec3) Mesh {
	return m.mapped(func(p, n Vec3) (Vec3, Vec3) { return p.Add(d), n })
}

// Scale scales by s per axis. Normals use the inverse scale; an odd number of
// negative factors mirrors the mesh, so the winding is reversed to keep
// triangles facing outward.
func (m Mesh) Scale(s Vec3) Mesh {
	out := m.mapped(func(p, n Vec3) (Vec3, Vec3) {
		return Vec3{p.X * s.X, p.Y * s.Y, p.Z * s.Z}, Vec3{n.X / s.X, n.Y / s.Y, n.Z / s.Z}.Norm()
	})
	if s.X*s.Y*s.Z < 0 {
		for i := 0; i < len(out.Indices); i += 3 {
			out.Indices[i+1], out.Indices[i+2] = out.Indices[i+2], out.Indices[i+1]
		}
	}
	return out
}

func rot(a float64, f func(v Vec3, c, s float64) Vec3) func(p, n Vec3) (Vec3, Vec3) {
	c, s := math.Cos(a), math.Sin(a)
	return func(p, n Vec3) (Vec3, Vec3) { return f(p, c, s), f(n, c, s) }
}

// RotateX, RotateY and RotateZ rotate right-handedly by a radians about the
// axis through the origin.
func (m Mesh) RotateX(a float64) Mesh {
	return m.mapped(rot(a, func(v Vec3, c, s float64) Vec3 { return Vec3{v.X, v.Y*c - v.Z*s, v.Y*s + v.Z*c} }))
}
func (m Mesh) RotateY(a float64) Mesh {
	return m.mapped(rot(a, func(v Vec3, c, s float64) Vec3 { return Vec3{v.X*c + v.Z*s, v.Y, -v.X*s + v.Z*c} }))
}
func (m Mesh) RotateZ(a float64) Mesh {
	return m.mapped(rot(a, func(v Vec3, c, s float64) Vec3 { return Vec3{v.X*c - v.Y*s, v.X*s + v.Y*c, v.Z} }))
}

// Bounds is the axis-aligned bounding box of the positions.
func (m Mesh) Bounds() (lo, hi Vec3) {
	lo = Vec3{math.Inf(1), math.Inf(1), math.Inf(1)}
	hi = Vec3{math.Inf(-1), math.Inf(-1), math.Inf(-1)}
	for _, p := range m.Positions {
		lo = Vec3{math.Min(lo.X, p.X), math.Min(lo.Y, p.Y), math.Min(lo.Z, p.Z)}
		hi = Vec3{math.Max(hi.X, p.X), math.Max(hi.Y, p.Y), math.Max(hi.Z, p.Z)}
	}
	return lo, hi
}

// Volume is the signed volume enclosed by the triangles: positive when they
// face outward.
func (m Mesh) Volume() float64 {
	v := 0.0
	for i := 0; i+2 < len(m.Indices); i += 3 {
		a, b, c := m.Positions[m.Indices[i]], m.Positions[m.Indices[i+1]], m.Positions[m.Indices[i+2]]
		v += a.Dot(b.Cross(c)) / 6
	}
	return v
}

// Closed reports whether the triangles form closed surfaces: every edge, by
// position (vertices are duplicated where normals differ), is used by exactly
// two triangles, once in each direction. Degenerate triangles are ignored.
func (m Mesh) Closed() bool {
	type key [3]int64
	q := func(p Vec3) key {
		return key{int64(math.Round(p.X * 1e6)), int64(math.Round(p.Y * 1e6)), int64(math.Round(p.Z * 1e6))}
	}
	type edge [2]key
	count := map[edge]int{}
	for i := 0; i+2 < len(m.Indices); i += 3 {
		k := [3]key{q(m.Positions[m.Indices[i]]), q(m.Positions[m.Indices[i+1]]), q(m.Positions[m.Indices[i+2]])}
		if k[0] == k[1] || k[1] == k[2] || k[0] == k[2] {
			continue
		}
		for j := 0; j < 3; j++ {
			count[edge{k[j], k[(j+1)%3]}]++
		}
	}
	for e, n := range count {
		if n != 1 || count[edge{e[1], e[0]}] != 1 {
			return false
		}
	}
	return len(count) > 0
}

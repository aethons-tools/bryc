package mesh

import "math"

const (
	ellipsoidSlices = 24
	ellipsoidStacks = 12
	cylinderSides   = 32
	tubeSides       = 12
)

// Ellipsoid is a closed ellipsoid centred on the origin.
func Ellipsoid(rx, ry, rz float64) Mesh {
	var m Mesh
	add := func(p Vec3) uint32 {
		m.Positions = append(m.Positions, p)
		m.Normals = append(m.Normals, Vec3{p.X / (rx * rx), p.Y / (ry * ry), p.Z / (rz * rz)}.Norm())
		return uint32(len(m.Positions) - 1)
	}
	top := add(Vec3{0, ry, 0})
	rings := make([][]uint32, 0, ellipsoidStacks-1)
	for st := 1; st < ellipsoidStacks; st++ {
		phi := math.Pi * float64(st) / ellipsoidStacks
		var ring []uint32
		for sl := 0; sl < ellipsoidSlices; sl++ {
			th := 2 * math.Pi * float64(sl) / ellipsoidSlices
			ring = append(ring, add(Vec3{rx * math.Sin(phi) * math.Cos(th), ry * math.Cos(phi), rz * math.Sin(phi) * math.Sin(th)}))
		}
		rings = append(rings, ring)
	}
	bottom := add(Vec3{0, -ry, 0})
	for sl := 0; sl < ellipsoidSlices; sl++ {
		j := (sl + 1) % ellipsoidSlices
		m.Indices = append(m.Indices, top, rings[0][j], rings[0][sl])
		last := rings[len(rings)-1]
		m.Indices = append(m.Indices, bottom, last[sl], last[j])
	}
	for r := 0; r+1 < len(rings); r++ {
		a, b := rings[r], rings[r+1]
		for sl := 0; sl < ellipsoidSlices; sl++ {
			j := (sl + 1) % ellipsoidSlices
			m.Indices = append(m.Indices, a[sl], a[j], b[j], a[sl], b[j], b[sl])
		}
	}
	if m.Volume() < 0 {
		flip(&m)
	}
	return m
}

// flip reverses every triangle's winding.
func flip(m *Mesh) {
	for i := 0; i < len(m.Indices); i += 3 {
		m.Indices[i+1], m.Indices[i+2] = m.Indices[i+2], m.Indices[i+1]
	}
}

// Cylinder is a closed cylinder of radius r along the Z axis, centred on the
// origin.
func Cylinder(r, length float64) Mesh {
	path := []Vec3{{0, 0, -length / 2}, {0, 0, length / 2}}
	return tube(path, func(float64) float64 { return r }, cylinderSides)
}

// Tube is a closed tube along path whose radius at arc-length fraction t is
// radius(t).
func Tube(path []Vec3, radius func(t float64) float64) Mesh {
	return tube(path, radius, tubeSides)
}

func tube(path []Vec3, radius func(float64) float64, sides int) Mesh {
	// Drop repeated points and measure arc length.
	var pts []Vec3
	for _, p := range path {
		if len(pts) == 0 || p.Sub(pts[len(pts)-1]).Len() > 1e-9 {
			pts = append(pts, p)
		}
	}
	if len(pts) < 2 {
		panic("mesh: tube needs two distinct points")
	}
	lengths := []float64{0}
	for i := 1; i < len(pts); i++ {
		lengths = append(lengths, lengths[i-1]+pts[i].Sub(pts[i-1]).Len())
	}
	total := lengths[len(lengths)-1]
	tangent := func(i int) Vec3 {
		a, b := pts[max(i-1, 0)], pts[min(i+1, len(pts)-1)]
		return b.Sub(a).Norm()
	}
	// Parallel-transported frame.
	t0 := tangent(0)
	ref := Vec3{0, 0, 1}
	if math.Abs(t0.Dot(ref)) > 0.9 {
		ref = Vec3{1, 0, 0}
	}
	nrm := ref.Sub(t0.Scale(ref.Dot(t0))).Norm()
	var m Mesh
	var rings [][]uint32
	for i, p := range pts {
		t := tangent(i)
		nrm = nrm.Sub(t.Scale(nrm.Dot(t))).Norm()
		bin := t.Cross(nrm)
		r := math.Max(radius(lengths[i]/total), 1e-3)
		var ring []uint32
		for s := 0; s < sides; s++ {
			a := 2 * math.Pi * float64(s) / float64(sides)
			d := nrm.Scale(math.Cos(a)).Add(bin.Scale(math.Sin(a)))
			m.Positions = append(m.Positions, p.Add(d.Scale(r)))
			m.Normals = append(m.Normals, d)
			ring = append(ring, uint32(len(m.Positions)-1))
		}
		rings = append(rings, ring)
	}
	for r := 0; r+1 < len(rings); r++ {
		a, b := rings[r], rings[r+1]
		for s := 0; s < sides; s++ {
			j := (s + 1) % sides
			m.Indices = append(m.Indices, a[s], a[j], b[j], a[s], b[j], b[s])
		}
	}
	// Caps: separate vertices with flat normals, fanned from a centre vertex.
	for _, end := range []struct {
		i   int
		dir float64
	}{{0, -1}, {len(pts) - 1, 1}} {
		n := tangent(end.i).Scale(end.dir)
		centre := uint32(len(m.Positions))
		m.Positions = append(m.Positions, pts[end.i])
		m.Normals = append(m.Normals, n)
		base := uint32(len(m.Positions))
		for _, v := range rings[end.i] {
			m.Positions = append(m.Positions, m.Positions[v])
			m.Normals = append(m.Normals, n)
		}
		for s := 0; s < sides; s++ {
			j := (s + 1) % sides
			if end.dir > 0 {
				m.Indices = append(m.Indices, centre, base+uint32(s), base+uint32(j))
			} else {
				m.Indices = append(m.Indices, centre, base+uint32(j), base+uint32(s))
			}
		}
	}
	if m.Volume() < 0 {
		flip(&m)
	}
	return m
}

package mesh

import (
	"math"
	"testing"
)

// cube is a closed, outward-facing unit cube built by hand, for testing the
// validation helpers.
func cube() Mesh {
	p := []Vec3{{0, 0, 0}, {1, 0, 0}, {1, 1, 0}, {0, 1, 0}, {0, 0, 1}, {1, 0, 1}, {1, 1, 1}, {0, 1, 1}}
	idx := []uint32{
		0, 2, 1, 0, 3, 2, // z=0, facing -z
		4, 5, 6, 4, 6, 7, // z=1, facing +z
		0, 1, 5, 0, 5, 4, // y=0
		3, 7, 6, 3, 6, 2, // y=1
		0, 4, 7, 0, 7, 3, // x=0
		1, 2, 6, 1, 6, 5, // x=1
	}
	n := make([]Vec3, len(p))
	for i := range n {
		n[i] = p[i].Sub(Vec3{0.5, 0.5, 0.5}).Norm()
	}
	return Mesh{Positions: p, Normals: n, Indices: idx}
}

func TestCubeClosedAndVolume(t *testing.T) {
	c := cube()
	if !c.Closed() {
		t.Error("cube not closed")
	}
	if v := c.Volume(); math.Abs(v-1) > 1e-9 {
		t.Errorf("volume = %v, want 1", v)
	}
}

func TestOpenMeshNotClosed(t *testing.T) {
	c := cube()
	c.Indices = c.Indices[:len(c.Indices)-3]
	if c.Closed() {
		t.Error("mesh with a missing triangle reported closed")
	}
}

func TestInsideOutHasNegativeVolume(t *testing.T) {
	c := cube()
	for i := 0; i < len(c.Indices); i += 3 {
		c.Indices[i+1], c.Indices[i+2] = c.Indices[i+2], c.Indices[i+1]
	}
	if c.Volume() >= 0 {
		t.Error("inside-out cube should have negative volume")
	}
}

func TestTransforms(t *testing.T) {
	c := cube().Translate(Vec3{-0.5, -0.5, -0.5})
	s := c.Scale(Vec3{2, 3, 4})
	if v := s.Volume(); math.Abs(v-24) > 1e-9 {
		t.Errorf("scaled volume = %v, want 24", v)
	}
	m := c.Scale(Vec3{-1, 1, 1}) // mirror: winding must be fixed up
	if m.Volume() <= 0 || !m.Closed() {
		t.Errorf("mirrored cube volume %v closed %v", m.Volume(), m.Closed())
	}
	r := c.RotateZ(math.Pi / 2)
	lo, hi := r.Bounds()
	if math.Abs(lo.X+0.5) > 1e-9 || math.Abs(hi.Y-0.5) > 1e-9 {
		t.Errorf("rotated bounds %v %v", lo, hi)
	}
	if p := (Mesh{Positions: []Vec3{{1, 0, 0}}, Normals: []Vec3{{1, 0, 0}}}).RotateZ(math.Pi / 2).Positions[0]; math.Abs(p.Y-1) > 1e-9 {
		t.Errorf("RotateZ(+90°) of +X = %v, want +Y (right-handed)", p)
	}
	if p := (Mesh{Positions: []Vec3{{0, 0, 1}}, Normals: []Vec3{{0, 0, 1}}}).RotateY(math.Pi / 2).Positions[0]; math.Abs(p.X-1) > 1e-9 {
		t.Errorf("RotateY(+90°) of +Z = %v, want +X", p)
	}
	if p := (Mesh{Positions: []Vec3{{0, 1, 0}}, Normals: []Vec3{{0, 1, 0}}}).RotateX(math.Pi / 2).Positions[0]; math.Abs(p.Z-1) > 1e-9 {
		t.Errorf("RotateX(+90°) of +Y = %v, want +Z", p)
	}
}

func TestAppend(t *testing.T) {
	var m Mesh
	m.Append(cube())
	m.Append(cube().Translate(Vec3{5, 0, 0}))
	if len(m.Positions) != 16 || len(m.Indices) != 72 || !m.Closed() || math.Abs(m.Volume()-2) > 1e-9 {
		t.Errorf("appended: %d positions, %d indices, closed %v, volume %v", len(m.Positions), len(m.Indices), m.Closed(), m.Volume())
	}
}

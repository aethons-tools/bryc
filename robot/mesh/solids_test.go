package mesh

import (
	"math"
	"testing"
)

func TestEllipsoid(t *testing.T) {
	m := Ellipsoid(3, 2, 1)
	if !m.Closed() {
		t.Fatal("not closed")
	}
	want := 4.0 / 3 * math.Pi * 3 * 2 * 1
	if v := m.Volume(); v <= 0 || math.Abs(v-want)/want > 0.03 {
		t.Errorf("volume %v, want ≈ %v", v, want)
	}
	lo, hi := m.Bounds()
	if math.Abs(hi.X-3) > 1e-9 || math.Abs(lo.Y+2) > 1e-9 || math.Abs(hi.Z-1) > 1e-9 {
		t.Errorf("bounds %v %v", lo, hi)
	}
}

func TestCylinder(t *testing.T) {
	m := Cylinder(2, 10)
	if !m.Closed() {
		t.Fatal("not closed")
	}
	want := math.Pi * 4 * 10
	if v := m.Volume(); v <= 0 || math.Abs(v-want)/want > 0.02 {
		t.Errorf("volume %v, want ≈ %v", v, want)
	}
	lo, hi := m.Bounds()
	if math.Abs(lo.Z+5) > 1e-9 || math.Abs(hi.Z-5) > 1e-9 {
		t.Errorf("z bounds %v..%v", lo.Z, hi.Z)
	}
}

func TestTube(t *testing.T) {
	straight := Tube([]Vec3{{0, 0, 0}, {10, 0, 0}}, func(float64) float64 { return 1 })
	if !straight.Closed() || straight.Volume() <= 0 {
		t.Fatalf("straight tube closed %v volume %v", straight.Closed(), straight.Volume())
	}
	if v, want := straight.Volume(), math.Pi*10; math.Abs(v-want)/want > 0.05 {
		t.Errorf("straight volume %v, want ≈ %v", v, want)
	}
	var arc []Vec3
	for i := 0; i <= 20; i++ {
		a := math.Pi * float64(i) / 20
		arc = append(arc, Vec3{10 * math.Cos(a), 10 * math.Sin(a), 0})
	}
	tapered := Tube(arc, func(t float64) float64 { return 2 * (1 - t) })
	if !tapered.Closed() || tapered.Volume() <= 0 {
		t.Errorf("tapered arc closed %v volume %v", tapered.Closed(), tapered.Volume())
	}
}

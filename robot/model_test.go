package robot

import (
	"bytes"
	"math"
	"testing"

	"github.com/aethons-tools/bryc/robot/mesh"
)

func nodeNames(s Spec) map[string]mesh.Mesh {
	out := map[string]mesh.Mesh{}
	for _, n := range Model(s).Nodes {
		out[n.Name] = n.Mesh
	}
	return out
}

func TestModelCoreParts(t *testing.T) {
	s := Resolve(Spec{}, 1)
	nodes := nodeNames(s)
	for _, name := range []string{"head", "neck", "shoulders", "chest-light"} {
		m, ok := nodes[name]
		if !ok {
			t.Fatalf("missing node %q", name)
		}
		if !m.Closed() || m.Volume() <= 0 {
			t.Errorf("%s: closed %v volume %v", name, m.Closed(), m.Volume())
		}
	}
}

func TestModelBoundsAndUnits(t *testing.T) {
	for seed := uint64(0); seed < 30; seed++ {
		for _, n := range Model(Resolve(Spec{}, seed)).Nodes {
			lo, hi := n.Mesh.Bounds()
			if lo.X < -0.5 || hi.X > 0.5 || lo.Y < -1e-9 || hi.Y > 1 || lo.Z < -0.3 || hi.Z > 0.3 {
				t.Fatalf("seed %d %s: bounds %v..%v outside the 1 m box", seed, n.Name, lo, hi)
			}
		}
	}
}

func TestModelHeadMatchesSilhouette(t *testing.T) {
	no := false
	for _, head := range HeadValues {
		s := Resolve(Spec{Head: head, Tall: &no}, 2)
		b := headLayout(s)
		h := nodeNames(s)["head"]
		lo, hi := h.Bounds()
		// The head's width in metres equals the 2D head's width at its widest
		// row (the top row for the inverted trapezoid, so include the ends).
		widest := 0.0
		for f := 0.0; f <= 1.0001; f += 0.05 {
			widest = math.Max(widest, 2*(b.CX()-leftEdge(head, b, b.Y+b.H*f)))
		}
		if got := (hi.X - lo.X) * 1000; math.Abs(got-widest) > 2 {
			t.Errorf("%s: model width %v, 2D widest %v", head, got, widest)
		}
		if math.Abs(hi.Z*1000-faceZ) > 1e-6 {
			t.Errorf("%s: front face at z %v, want %v", head, hi.Z*1000, faceZ)
		}
	}
}

func TestModelDeterministic(t *testing.T) {
	s := Resolve(Spec{}, 9)
	var a, b bytes.Buffer
	if err := WriteGLB(&a, s); err != nil {
		t.Fatal(err)
	}
	WriteGLB(&b, s)
	if !bytes.Equal(a.Bytes(), b.Bytes()) {
		t.Error("same spec gave different .glb bytes")
	}
}

func TestLinear(t *testing.T) {
	c := linear(parseHex("#ffffff"), 1)
	d := linear(parseHex("#808080"), 0.5)
	if c != [4]float64{1, 1, 1, 1} || math.Abs(d[0]-0.2158605) > 1e-6 || d[3] != 0.5 {
		t.Errorf("linear white %v grey %v", c, d)
	}
}

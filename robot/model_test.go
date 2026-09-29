package robot

import (
	"bytes"
	"fmt"
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

func TestModelFaceParts(t *testing.T) {
	no := false
	cases := []struct {
		spec Spec
		want []string
	}{
		{Spec{Eyes: "round", EyeCount: "2", EyeStyle: "bright", Mouth: "line"}, []string{"eye-0", "eye-1", "mouth"}},
		{Spec{Eyes: "focused", EyeCount: "4", EyeStyle: "glower", Mouth: "grille"}, []string{"eye-3", "eye-ring-3", "eye-core-3", "mouth", "mouth-bar-0", "mouth-bar-4"}},
		{Spec{Eyes: "oval", EyeCount: "1", EyeStyle: "dead", Mouth: "slot"}, []string{"eye-0", "mouth"}},
		{Spec{Eyes: "visor", Mouth: "jaw"}, []string{"visor", "visor-bar", "jaw", "jaw-bolt-0", "jaw-bolt-1"}},
	}
	for _, c := range cases {
		c.spec.Eyelashes = &no
		nodes := nodeNames(Resolve(c.spec, 3))
		for _, name := range c.want {
			m, ok := nodes[name]
			if !ok {
				t.Errorf("%+v: missing %q", c.spec, name)
				continue
			}
			if !m.Closed() || m.Volume() <= 0 {
				t.Errorf("%s: closed %v volume %v", name, m.Closed(), m.Volume())
			}
			// Face parts reach in front of the face.
			if _, hi := m.Bounds(); name != "jaw" && hi.Z*1000 <= faceZ {
				t.Errorf("%s: front at z %v, not in front of the face (%v)", name, hi.Z*1000, faceZ)
			}
		}
	}
}

func TestModelEyesFollowLayout(t *testing.T) {
	s := Resolve(Spec{Eyes: "round", EyeCount: "6", EyeStyle: "dead"}, 4)
	nodes := nodeNames(s)
	for i, p := range roundEyeCenters(s, headLayout(s)) {
		lo, hi := nodes[fmt.Sprintf("eye-%d", i)].Bounds()
		cx, cy := (lo.X+hi.X)/2*1000+500, virtual-(lo.Y+hi.Y)/2*1000
		if math.Abs(cx-p[0]) > 0.5 || math.Abs(cy-p[1]) > 0.5 {
			t.Errorf("eye %d centre (%v,%v), 2D (%v,%v)", i, cx, cy, p[0], p[1])
		}
	}
}

package robot

import (
	"bytes"
	"fmt"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

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

func TestModelFocusedEyeTilt(t *testing.T) {
	no := false
	s := Resolve(Spec{Eyes: "focused", EyeCount: "2", EyeStyle: "dead", Eyelashes: &no}, 5)
	nodes := nodeNames(s)
	// yAtX returns the Y of the vertex with the largest (or smallest) X.
	yAtX := func(m mesh.Mesh, max bool) float64 {
		best := m.Positions[0]
		for _, p := range m.Positions {
			if (max && p.X > best.X) || (!max && p.X < best.X) {
				best = p
			}
		}
		return best.Y
	}
	// Inner ends tilt down: the left eye's right end and the right eye's left end.
	l, r := nodes["eye-0"], nodes["eye-1"]
	if yAtX(l, true) >= yAtX(l, false) {
		t.Errorf("left eye: inner (right) end not lower than outer end")
	}
	if yAtX(r, false) >= yAtX(r, true) {
		t.Errorf("right eye: inner (left) end not lower than outer end")
	}
}

func TestModelEveryValueBuilds(t *testing.T) {
	for _, facet := range Facets() {
		values := facet.Values
		switch facet.Kind {
		case KindBool:
			values = []string{"true", "false"}
		case KindRange:
			values = []string{strconv.Itoa(*facet.Min), "0", strconv.Itoa(*facet.Max)}
		case KindColor:
			continue
		}
		for _, v := range values {
			req, err := ParseQuery(url.Values{facet.Name: {v}})
			if err != nil {
				t.Fatal(err)
			}
			for _, n := range Model(Resolve(req.Spec, 11)).Nodes {
				if !n.Mesh.Closed() || n.Mesh.Volume() <= 0 {
					t.Errorf("%s=%s: node %s closed %v volume %v", facet.Name, v, n.Name, n.Mesh.Closed(), n.Mesh.Volume())
				}
			}
		}
	}
}

func TestModelHardwareParts(t *testing.T) {
	yes := true
	s := Resolve(Spec{Eyes: "round", EyeCount: "2", Eyelashes: &yes, Rivets: &yes, Panels: &yes, Blush: &yes,
		Ears: "dials", Antenna: "double"}, 5)
	nodes := nodeNames(s)
	for _, name := range []string{"rivet-0", "rivet-3", "seam-0", "seam-1", "blush-0", "blush-1", "lash-0", "lash-1",
		"ear-0", "ear-1", "ear-cap-0", "ear-tick-1", "antenna-stem-0", "antenna-stem-1", "antenna-ball-1", "antenna-base-0"} {
		if _, ok := nodes[name]; !ok {
			t.Errorf("missing %q", name)
		}
	}
	s2 := Resolve(Spec{Ears: "bolts", Antenna: "bolt"}, 5)
	for _, name := range []string{"ear-0", "ear-dot-1", "antenna-bolt", "antenna-base-0"} {
		if _, ok := nodeNames(s2)[name]; !ok {
			t.Errorf("missing %q", name)
		}
	}
}

func TestModelBudget(t *testing.T) {
	for seed := uint64(0); seed < 20; seed++ {
		s := Resolve(Spec{}, seed)
		start := time.Now()
		var buf bytes.Buffer
		if err := WriteGLB(&buf, s); err != nil {
			t.Fatal(err)
		}
		if d := time.Since(start); d > 100*time.Millisecond {
			t.Errorf("seed %d: built in %v, budget 100ms", seed, d)
		}
		if buf.Len() > 1<<20 {
			t.Errorf("seed %d: %d bytes, budget 1 MB", seed, buf.Len())
		}
	}
}

func TestModelGolden(t *testing.T) {
	for name, spec := range map[string]Spec{
		"model-seed-1": Resolve(Spec{}, 1),
		"model-seed-2": Resolve(Spec{}, 2),
		"model-seed-3": Resolve(Spec{}, 3),
	} {
		var buf bytes.Buffer
		if err := WriteGLB(&buf, spec); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join("testdata", name+".glb")
		if *update {
			if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
				t.Fatal(err)
			}
			continue
		}
		want, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("%v (generate with: go test ./robot -update -run TestModelGolden)", err)
		}
		if !bytes.Equal(buf.Bytes(), want) {
			t.Errorf("%s changed; if intended run: go test ./robot -update -run TestModelGolden", name)
		}
	}
}

func TestModelHardwarePlacement(t *testing.T) {
	yes := true
	for _, head := range HeadValues {
		for _, tall := range []bool{false, true} {
			tall := tall
			s := Resolve(Spec{Head: head, Tall: &tall, Ears: "dials", Panels: &yes}, 4)
			b := headLayout(s)
			nodes := nodeNames(s)
			canvasX := func(v float64) float64 { return v*1000 + 500 }
			y := earY(b)
			edge := leftEdge(s.Head, b, y)
			mirror := 2*b.CX() - edge
			lo, hi := nodes["ear-0"].Bounds()
			if in := canvasX(hi.X) - edge; in < 10 {
				t.Errorf("%s tall=%v: left dial only %.1f inside the head", head, tall, in)
			}
			lo, hi = nodes["ear-1"].Bounds()
			if in := mirror - canvasX(lo.X); in < 10 {
				t.Errorf("%s tall=%v: right dial only %.1f inside the head", head, tall, in)
			}
			for i, seam := range panelSeams(s, b) {
				lo, hi := nodes[fmt.Sprintf("seam-%d", i)].Bounds()
				e := leftEdge(s.Head, b, seam.y)
				if canvasX(lo.X) < e+12 || canvasX(hi.X) > 2*b.CX()-e-12 {
					t.Errorf("%s tall=%v seam-%d: x %.1f..%.1f outside [%.1f, %.1f]", head, tall, i,
						canvasX(lo.X), canvasX(hi.X), e+12, 2*b.CX()-e-12)
				}
			}
		}
	}
}

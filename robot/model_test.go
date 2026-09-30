package robot

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
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
	for _, name := range []string{"head"} {
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
			if lo.X < -0.5 || hi.X > 0.5 || lo.Y < -1e-9 || hi.Y > 1 || lo.Z < -0.3 || hi.Z > 0.4 { // the hands float in front
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
		if err := compareGLB(buf.Bytes(), want); err != nil {
			t.Errorf("%s changed (%v); if intended run: go test ./robot -update -run TestModelGolden", name, err)
		}
	}
}

// glbTol is the absolute tolerance (metres) for float data: arm64 fuses
// multiply-adds where amd64 does not, so float32 low bits differ by platform.
const glbTol = 1e-6

// splitGLB returns the JSON chunk decoded generically and the BIN chunk.
func splitGLB(b []byte) (map[string]any, []byte, error) {
	le := binary.LittleEndian
	if len(b) < 28 || le.Uint32(b[0:]) != 0x46546C67 || int(le.Uint32(b[8:])) != len(b) {
		return nil, nil, fmt.Errorf("bad GLB header")
	}
	jl := int(le.Uint32(b[12:]))
	if le.Uint32(b[16:]) != 0x4E4F534A || 20+jl+8 > len(b) {
		return nil, nil, fmt.Errorf("bad JSON chunk")
	}
	var doc map[string]any
	if err := json.Unmarshal(b[20:20+jl], &doc); err != nil {
		return nil, nil, err
	}
	rest := b[20+jl:]
	if le.Uint32(rest[4:]) != 0x004E4942 || int(le.Uint32(rest))+8 != len(rest) {
		return nil, nil, fmt.Errorf("bad BIN chunk")
	}
	return doc, rest[8:], nil
}

// jsonNear is deep equality where numbers may differ by 1e-9 (material colors
// come from float64 pow/multiply chains that differ in the last digit across
// architectures); everything else must match exactly.
func jsonNear(a, b any) bool {
	switch x := a.(type) {
	case float64:
		y, ok := b.(float64)
		return ok && math.Abs(x-y) <= 1e-9
	case []any:
		y, ok := b.([]any)
		if !ok || len(x) != len(y) {
			return false
		}
		for i := range x {
			if !jsonNear(x[i], y[i]) {
				return false
			}
		}
		return true
	case map[string]any:
		y, ok := b.(map[string]any)
		if !ok || len(x) != len(y) {
			return false
		}
		for k, v := range x {
			w, ok := y[k]
			if !ok || !jsonNear(v, w) {
				return false
			}
		}
		return true
	}
	return reflect.DeepEqual(a, b)
}

// compareGLB requires got and want to match: identical JSON (numbers within 1e-9) except accessor
// min/max, identical index data, and float data within glbTol.
func compareGLB(got, want []byte) error {
	gd, gb, err := splitGLB(got)
	if err != nil {
		return err
	}
	wd, wb, err := splitGLB(want)
	if err != nil {
		return err
	}
	if len(gb) != len(wb) {
		return fmt.Errorf("BIN length %d, want %d", len(gb), len(wb))
	}
	ga, _ := gd["accessors"].([]any)
	wa, _ := wd["accessors"].([]any)
	if len(ga) != len(wa) {
		return fmt.Errorf("%d accessors, want %d", len(ga), len(wa))
	}
	type acc struct{ min, max []any }
	strip := func(as []any) []acc {
		out := make([]acc, len(as))
		for i, a := range as {
			m := a.(map[string]any)
			out[i].min, _ = m["min"].([]any)
			out[i].max, _ = m["max"].([]any)
			delete(m, "min")
			delete(m, "max")
		}
		return out
	}
	gm, wm := strip(ga), strip(wa)
	if !jsonNear(gd, wd) {
		return fmt.Errorf("JSON differs")
	}
	near := func(a, b []any) bool {
		if len(a) != len(b) {
			return false
		}
		for i := range a {
			x, _ := a[i].(float64)
			y, _ := b[i].(float64)
			if math.Abs(x-y) > glbTol {
				return false
			}
		}
		return true
	}
	views := wd["bufferViews"].([]any)
	le := binary.LittleEndian
	for i, a := range wa {
		m := a.(map[string]any)
		if !near(gm[i].min, wm[i].min) || !near(gm[i].max, wm[i].max) {
			return fmt.Errorf("accessor %d min/max differ", i)
		}
		v := views[int(m["bufferView"].(float64))].(map[string]any)
		off, n := int(v["byteOffset"].(float64)), int(v["byteLength"].(float64))
		g, w := gb[off:off+n], wb[off:off+n]
		if m["componentType"].(float64) != 5126 {
			if !bytes.Equal(g, w) {
				return fmt.Errorf("accessor %d index data differs", i)
			}
			continue
		}
		for k := 0; k+4 <= n; k += 4 {
			x := math.Float32frombits(le.Uint32(g[k:]))
			y := math.Float32frombits(le.Uint32(w[k:]))
			if math.Abs(float64(x)-float64(y)) > glbTol {
				return fmt.Errorf("accessor %d float %d: %v, want %v", i, k/4, x, y)
			}
		}
	}
	return nil
}

func TestModelJawFront(t *testing.T) {
	yes := true
	for _, head := range HeadValues {
		for _, tall := range []bool{false, true} {
			for _, expr := range ExpressionValues {
				for face := FaceMin; face <= FaceMax; face++ {
					tall, face := tall, face
					s := Resolve(Spec{Head: head, Tall: &tall, Expression: expr, Face: &face, Mouth: "jaw", Blush: &yes}, 6)
					nodes := nodeNames(s)
					_, jaw := nodes["jaw"].Bounds()
					for name, m := range nodes {
						if !strings.HasPrefix(name, "blush-") {
							continue
						}
						if _, hi := m.Bounds(); hi.Z*1000 <= jaw.Z*1000 {
							t.Errorf("%s tall=%v %s face=%d: %s front %v not in front of jaw %v", head, tall, expr, face, name, hi.Z*1000, jaw.Z*1000)
						}
					}
				}
			}
		}
	}
}

func TestModelVisorClearsJaw(t *testing.T) {
	for _, head := range HeadValues {
		for _, tall := range []bool{false, true} {
			for face := FaceMin; face <= FaceMax; face++ {
				tall, face := tall, face
				s := Resolve(Spec{Head: head, Tall: &tall, Face: &face, Eyes: "visor", Mouth: "jaw"}, 6)
				nodes := nodeNames(s)
				_, jaw := nodes["jaw"].Bounds()
				for _, name := range []string{"visor", "visor-bar"} {
					if _, hi := nodes[name].Bounds(); hi.Z*1000 <= jaw.Z*1000 {
						t.Errorf("%s tall=%v face=%d: %s front %v not in front of jaw %v", head, tall, face, name, hi.Z*1000, jaw.Z*1000)
					}
				}
			}
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

package glb

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"math"
	"testing"

	"github.com/aethons-tools/bryc/robot/mesh"
)

type gltf struct {
	Asset  struct{ Version, Generator string }
	Scene  int
	Scenes []struct{ Nodes []int }
	Nodes  []struct {
		Name     string
		Mesh     *int
		Children []int
	}
	Meshes []struct {
		Primitives []struct {
			Attributes map[string]int
			Indices    int
			Material   int
			Mode       int
		}
	}
	Materials []struct {
		Name                 string
		PbrMetallicRoughness struct {
			BaseColorFactor []float64
			MetallicFactor  float64
			RoughnessFactor float64
		}
		EmissiveFactor []float64
		AlphaMode      string
	}
	Accessors []struct {
		BufferView    int
		ComponentType int
		Count         int
		Type          string
		Min, Max      []float64
	}
	BufferViews []struct {
		Buffer, ByteOffset, ByteLength, Target int
	}
	Buffers []struct{ ByteLength int }
}

// parse checks the GLB container and returns the JSON and binary chunks.
func parse(t *testing.T, b []byte) (gltf, []byte) {
	t.Helper()
	le := binary.LittleEndian
	if len(b) < 20 || le.Uint32(b[0:]) != 0x46546C67 || le.Uint32(b[4:]) != 2 || int(le.Uint32(b[8:])) != len(b) {
		t.Fatalf("bad GLB header")
	}
	jlen := int(le.Uint32(b[12:]))
	if le.Uint32(b[16:]) != 0x4E4F534A || jlen%4 != 0 {
		t.Fatalf("bad JSON chunk")
	}
	var g gltf
	if err := json.Unmarshal(b[20:20+jlen], &g); err != nil {
		t.Fatal(err)
	}
	rest := b[20+jlen:]
	blen := int(le.Uint32(rest[0:]))
	if le.Uint32(rest[4:]) != 0x004E4942 || blen%4 != 0 || len(rest) != 8+blen {
		t.Fatalf("bad BIN chunk")
	}
	return g, rest[8:]
}

func testScene() Scene {
	box, _ := mesh.Extrude([]mesh.Vec2{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}}, 1, 0.1)
	ball := mesh.Ellipsoid(1, 1, 1)
	red := Material{Name: "red", Color: [4]float64{1, 0, 0, 1}, Roughness: 0.5}
	glow := Material{Name: "glow", Color: [4]float64{0, 1, 0, 1}, Roughness: 0.4, Emissive: [3]float64{0, 1, 0}}
	return Scene{Nodes: []Node{
		{Name: "box", Mesh: box, Material: red},
		{Name: "ball", Mesh: ball, Material: glow},
		{Name: "box2", Mesh: box.Translate(mesh.Vec3{X: 3}), Material: red},
	}}
}

func TestEncodeValidGLB(t *testing.T) {
	s := testScene()
	b, err := Encode(s)
	if err != nil {
		t.Fatal(err)
	}
	g, bin := parse(t, b)
	if g.Asset.Version != "2.0" || g.Asset.Generator != "bryc" {
		t.Errorf("asset %+v", g.Asset)
	}
	if len(g.Buffers) != 1 || g.Buffers[0].ByteLength != len(bin) {
		t.Errorf("buffers %+v, bin %d", g.Buffers, len(bin))
	}
	root := g.Nodes[g.Scenes[g.Scene].Nodes[0]]
	if root.Name != "robot" || len(root.Children) != 3 {
		t.Errorf("root %+v", root)
	}
	if len(g.Materials) != 2 {
		t.Errorf("%d materials, want 2 (red deduplicated)", len(g.Materials))
	}
	if g.Materials[1].EmissiveFactor[1] != 1 {
		t.Errorf("glow emissive %v", g.Materials[1].EmissiveFactor)
	}
	for i, child := range root.Children {
		n := g.Nodes[child]
		if n.Name != s.Nodes[i].Name || n.Mesh == nil {
			t.Fatalf("node %d = %+v", i, n)
		}
		p := g.Meshes[*n.Mesh].Primitives[0]
		if p.Mode != 4 || p.Material < 0 || p.Material >= len(g.Materials) {
			t.Errorf("primitive %+v", p)
		}
		pos := g.Accessors[p.Attributes["POSITION"]]
		nrm := g.Accessors[p.Attributes["NORMAL"]]
		idx := g.Accessors[p.Indices]
		if pos.Count != len(s.Nodes[i].Mesh.Positions) || nrm.Count != pos.Count || idx.Count != len(s.Nodes[i].Mesh.Indices) || idx.Count%3 != 0 {
			t.Errorf("node %d counts pos %d nrm %d idx %d", i, pos.Count, nrm.Count, idx.Count)
		}
		if pos.ComponentType != 5126 || pos.Type != "VEC3" || idx.ComponentType != 5125 || idx.Type != "SCALAR" {
			t.Errorf("node %d accessor types", i)
		}
		lo, hi := s.Nodes[i].Mesh.Bounds()
		if math.Abs(pos.Min[0]-lo.X) > 1e-6 || math.Abs(pos.Max[2]-hi.Z) > 1e-6 {
			t.Errorf("node %d min/max %v %v, want %v %v", i, pos.Min, pos.Max, lo, hi)
		}
		for _, a := range []int{p.Attributes["POSITION"], p.Attributes["NORMAL"], p.Indices} {
			v := g.BufferViews[g.Accessors[a].BufferView]
			if v.ByteOffset%4 != 0 || v.ByteOffset+v.ByteLength > len(bin) {
				t.Errorf("buffer view %+v out of range or unaligned", v)
			}
		}
		// Indices in range.
		v := g.BufferViews[idx.BufferView]
		for k := 0; k < idx.Count; k++ {
			if int(binary.LittleEndian.Uint32(bin[v.ByteOffset+4*k:])) >= pos.Count {
				t.Fatalf("node %d index out of range", i)
			}
		}
	}
}

func TestEncodeDeterministic(t *testing.T) {
	a, _ := Encode(testScene())
	b, _ := Encode(testScene())
	if !bytes.Equal(a, b) {
		t.Error("same scene encoded differently")
	}
}

func TestEncodeBlendAndErrors(t *testing.T) {
	s := testScene()
	s.Nodes[0].Material = Material{Name: "blush", Color: [4]float64{1, 0, 1, 0.6}, Roughness: 0.45, Blend: true}
	b, err := Encode(s)
	if err != nil {
		t.Fatal(err)
	}
	g, _ := parse(t, b)
	if g.Materials[0].AlphaMode != "BLEND" {
		t.Errorf("alphaMode %q", g.Materials[0].AlphaMode)
	}
	if _, err := Encode(Scene{}); err == nil {
		t.Error("empty scene should fail")
	}
	if _, err := Encode(Scene{Nodes: []Node{{Name: "empty"}}}); err == nil {
		t.Error("empty mesh should fail")
	}
}

// Package glb writes scenes of coloured meshes as binary glTF 2.0 (.glb).
// It knows nothing about robots.
package glb

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"github.com/aethons-tools/bryc/robot/mesh"
)

// Material is a glTF metallic-roughness material. Colours are linear, 0..1.
type Material struct {
	Name                string
	Color               [4]float64
	Metallic, Roughness float64
	Emissive            [3]float64
	Blend               bool // alpha blending; otherwise opaque
}

// Node is one named mesh with its material.
type Node struct {
	Name     string
	Mesh     mesh.Mesh
	Material Material
}

// Scene is a flat list of nodes, written as children of a root node "robot".
type Scene struct{ Nodes []Node }

type jsonDoc struct {
	Asset       jsonAsset        `json:"asset"`
	Scene       int              `json:"scene"`
	Scenes      []jsonScene      `json:"scenes"`
	Nodes       []jsonNode       `json:"nodes"`
	Meshes      []jsonMesh       `json:"meshes"`
	Materials   []jsonMaterial   `json:"materials"`
	Accessors   []jsonAccessor   `json:"accessors"`
	BufferViews []jsonBufferView `json:"bufferViews"`
	Buffers     []jsonBuffer     `json:"buffers"`
}
type jsonAsset struct {
	Version   string `json:"version"`
	Generator string `json:"generator"`
}
type jsonScene struct {
	Nodes []int `json:"nodes"`
}
type jsonNode struct {
	Name     string `json:"name"`
	Mesh     *int   `json:"mesh,omitempty"`
	Children []int  `json:"children,omitempty"`
}
type jsonMesh struct {
	Name       string          `json:"name"`
	Primitives []jsonPrimitive `json:"primitives"`
}
type jsonPrimitive struct {
	Attributes jsonAttributes `json:"attributes"`
	Indices    int            `json:"indices"`
	Material   int            `json:"material"`
	Mode       int            `json:"mode"`
}
type jsonAttributes struct {
	Position int `json:"POSITION"`
	Normal   int `json:"NORMAL"`
}
type jsonMaterial struct {
	Name                 string    `json:"name"`
	PbrMetallicRoughness jsonPBR   `json:"pbrMetallicRoughness"`
	EmissiveFactor       []float64 `json:"emissiveFactor,omitempty"`
	AlphaMode            string    `json:"alphaMode,omitempty"`
}
type jsonPBR struct {
	BaseColorFactor []float64 `json:"baseColorFactor"`
	MetallicFactor  float64   `json:"metallicFactor"`
	RoughnessFactor float64   `json:"roughnessFactor"`
}
type jsonAccessor struct {
	BufferView    int       `json:"bufferView"`
	ComponentType int       `json:"componentType"`
	Count         int       `json:"count"`
	Type          string    `json:"type"`
	Min           []float64 `json:"min,omitempty"`
	Max           []float64 `json:"max,omitempty"`
}
type jsonBufferView struct {
	Buffer     int `json:"buffer"`
	ByteOffset int `json:"byteOffset"`
	ByteLength int `json:"byteLength"`
	Target     int `json:"target"`
}
type jsonBuffer struct {
	ByteLength int `json:"byteLength"`
}

const (
	componentFloat  = 5126
	componentUint32 = 5125
	targetArray     = 34962
	targetElement   = 34963
)

// Encode writes s as a .glb file.
func Encode(s Scene) ([]byte, error) {
	if len(s.Nodes) == 0 {
		return nil, errors.New("glb: empty scene")
	}
	doc := jsonDoc{
		Asset:  jsonAsset{Version: "2.0", Generator: "bryc"},
		Scenes: []jsonScene{{Nodes: []int{0}}},
		Nodes:  []jsonNode{{Name: "robot"}},
	}
	var bin bytes.Buffer
	view := func(data []byte, target int) int {
		for bin.Len()%4 != 0 {
			bin.WriteByte(0)
		}
		doc.BufferViews = append(doc.BufferViews, jsonBufferView{ByteOffset: bin.Len(), ByteLength: len(data), Target: target})
		bin.Write(data)
		return len(doc.BufferViews) - 1
	}
	vec3s := func(vs []mesh.Vec3, withBounds bool) int {
		buf := make([]byte, 12*len(vs))
		lo := []float64{math.Inf(1), math.Inf(1), math.Inf(1)}
		hi := []float64{math.Inf(-1), math.Inf(-1), math.Inf(-1)}
		for i, v := range vs {
			for k, c := range []float64{v.X, v.Y, v.Z} {
				f := float32(c)
				binary.LittleEndian.PutUint32(buf[12*i+4*k:], math.Float32bits(f))
				lo[k], hi[k] = math.Min(lo[k], float64(f)), math.Max(hi[k], float64(f))
			}
		}
		a := jsonAccessor{BufferView: view(buf, targetArray), ComponentType: componentFloat, Count: len(vs), Type: "VEC3"}
		if withBounds {
			a.Min, a.Max = lo, hi
		}
		doc.Accessors = append(doc.Accessors, a)
		return len(doc.Accessors) - 1
	}
	materials := map[Material]int{} // lookups only; order comes from doc.Materials
	for _, n := range s.Nodes {
		if len(n.Mesh.Positions) == 0 || len(n.Mesh.Indices) == 0 {
			return nil, fmt.Errorf("glb: node %q has an empty mesh", n.Name)
		}
		mat, ok := materials[n.Material]
		if !ok {
			m := n.Material
			jm := jsonMaterial{Name: m.Name, PbrMetallicRoughness: jsonPBR{
				BaseColorFactor: m.Color[:], MetallicFactor: m.Metallic, RoughnessFactor: m.Roughness}}
			if m.Emissive != [3]float64{} {
				jm.EmissiveFactor = m.Emissive[:]
			}
			if m.Blend {
				jm.AlphaMode = "BLEND"
			}
			doc.Materials = append(doc.Materials, jm)
			mat = len(doc.Materials) - 1
			materials[n.Material] = mat
		}
		pos := vec3s(n.Mesh.Positions, true)
		nrm := vec3s(n.Mesh.Normals, false)
		ibuf := make([]byte, 4*len(n.Mesh.Indices))
		for i, v := range n.Mesh.Indices {
			binary.LittleEndian.PutUint32(ibuf[4*i:], v)
		}
		doc.Accessors = append(doc.Accessors, jsonAccessor{BufferView: view(ibuf, targetElement), ComponentType: componentUint32, Count: len(n.Mesh.Indices), Type: "SCALAR"})
		idx := len(doc.Accessors) - 1
		doc.Meshes = append(doc.Meshes, jsonMesh{Name: n.Name, Primitives: []jsonPrimitive{{
			Attributes: jsonAttributes{Position: pos, Normal: nrm}, Indices: idx, Material: mat, Mode: 4}}})
		meshIndex := len(doc.Meshes) - 1
		doc.Nodes = append(doc.Nodes, jsonNode{Name: n.Name, Mesh: &meshIndex})
		doc.Nodes[0].Children = append(doc.Nodes[0].Children, len(doc.Nodes)-1)
	}
	for bin.Len()%4 != 0 {
		bin.WriteByte(0)
	}
	doc.Buffers = []jsonBuffer{{ByteLength: bin.Len()}}
	js, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}
	for len(js)%4 != 0 {
		js = append(js, ' ')
	}
	var out bytes.Buffer
	le := binary.LittleEndian
	total := 12 + 8 + len(js) + 8 + bin.Len()
	binary.Write(&out, le, [3]uint32{0x46546C67, 2, uint32(total)})
	binary.Write(&out, le, [2]uint32{uint32(len(js)), 0x4E4F534A})
	out.Write(js)
	binary.Write(&out, le, [2]uint32{uint32(bin.Len()), 0x004E4942})
	out.Write(bin.Bytes())
	return out.Bytes(), nil
}

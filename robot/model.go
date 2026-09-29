package robot

import (
	"fmt"
	"image/color"
	"io"
	"math"

	"github.com/aethons-tools/bryc/robot/glb"
	"github.com/aethons-tools/bryc/robot/mesh"
)

// The 3D model is built in "up canvas" units: x is canvas x−500, y is
// 1000−canvas y (so up is +y) and z points out of the face. Every node mesh is
// scaled by 0.001 at the end, so the model is about one metre tall.
const (
	headDepth                    = 300.0
	faceZ                        = headDepth / 2
	headBevel                    = 12.0
	neckDepth, neckBevel         = 140.0, 8.0
	shoulderDepth, shoulderBevel = 360.0, 12.0
)

type modelBuilder struct {
	s     Spec
	b     Layout
	mat   materials
	nodes []glb.Node
}

// add appends a named node; an empty mesh is a programming error.
func (m *modelBuilder) add(name string, me mesh.Mesh, mat glb.Material) {
	if len(me.Indices) == 0 {
		panic("robot: empty mesh for node " + name)
	}
	m.nodes = append(m.nodes, glb.Node{Name: name, Mesh: me, Material: mat})
}

// up converts canvas coordinates (y down) to model coordinates (y up).
func up(x, y float64) mesh.Vec2 { return mesh.Vec2{X: x - 500, Y: virtual - y} }

func upPoly(p []mesh.Vec2) []mesh.Vec2 {
	out := make([]mesh.Vec2, len(p))
	for i, v := range p {
		out[i] = up(v.X, v.Y)
	}
	return out
}

func upAt(x, y, z float64) mesh.Vec3 {
	v := up(x, y)
	return mesh.Vec3{X: v.X, Y: v.Y, Z: z}
}

func mustExtrude(poly []mesh.Vec2, depth, bevel float64) mesh.Mesh {
	me, err := mesh.Extrude(poly, depth, bevel)
	if err != nil {
		panic(fmt.Sprintf("robot: extrude: %v", err))
	}
	return me
}

// materials are the model's shared materials.
type materials struct {
	body, jaw, neck, antennaBase, accent, metal, screen, glow, ink, blush glb.Material
}

// linear converts an sRGB color to linear RGBA with the given alpha.
func linear(c color.NRGBA, alpha float64) [4]float64 {
	f := func(u uint8) float64 {
		v := float64(u) / 255
		if v <= 0.04045 {
			return v / 12.92
		}
		return math.Pow((v+0.055)/1.055, 2.4)
	}
	return [4]float64{f(c.R), f(c.G), f(c.B), alpha}
}

func materialsFor(s Spec) materials {
	body := parseHex(s.Body)
	accent := parseHex(s.Accent)
	glow := linear(parseHex(s.Glow), 1)
	return materials{
		body:        glb.Material{Name: "body", Color: linear(body, 1), Roughness: 0.45},
		jaw:         glb.Material{Name: "jaw", Color: linear(shade(body, jawShade), 1), Roughness: 0.45},
		neck:        glb.Material{Name: "neck", Color: linear(shade(body, 0.7), 1), Roughness: 0.45},
		antennaBase: glb.Material{Name: "antenna-base", Color: linear(shade(body, 0.8), 1), Roughness: 0.45},
		accent:      glb.Material{Name: "accent", Color: linear(accent, 1), Roughness: 0.45},
		metal:       glb.Material{Name: "metal", Color: linear(metal, 1), Metallic: 0.8, Roughness: 0.3},
		screen:      glb.Material{Name: "screen", Color: linear(screen, 1), Roughness: 0.15},
		glow:        glb.Material{Name: "glow", Color: glow, Roughness: 0.4, Emissive: [3]float64{glow[0], glow[1], glow[2]}},
		ink:         glb.Material{Name: "ink", Color: linear(ink, 1), Roughness: 0.5},
		blush:       glb.Material{Name: "blush", Color: linear(accent, 0.6), Roughness: 0.45, Blend: true},
	}
}

// Model builds the 3D robot for a resolved spec.
func Model(s Spec) glb.Scene {
	m := &modelBuilder{s: s, b: headLayout(s), mat: materialsFor(s)}

	m.add("head", mustExtrude(upPoly(headOutline(s, m.b)), headDepth, headBevel), m.mat.body)

	n := neckRect(m.b)
	m.add("neck", mustExtrude(upPoly(roundedRectOutline(n.x, n.y, n.w, n.h, 0)), neckDepth, neckBevel), m.mat.neck)

	shoulders := 0
	if s.Shoulders != nil {
		shoulders = *s.Shoulders
	}
	m.add("shoulders", mustExtrude(upPoly(shoulderOutline(shoulders)), shoulderDepth, shoulderBevel), m.mat.body)

	c := chestLight(m.b)
	m.add("chest-light", mesh.Ellipsoid(c.r, c.r, 15).Translate(upAt(c.x, c.y, shoulderDepth/2)), m.mat.accent)

	addFace(m)
	addHardware(m)

	k := mesh.Vec3{X: 0.001, Y: 0.001, Z: 0.001}
	for i := range m.nodes {
		m.nodes[i].Mesh = m.nodes[i].Mesh.Scale(k)
	}
	return glb.Scene{Nodes: m.nodes}
}

// WriteGLB writes the spec's 3D model as a binary glTF.
func WriteGLB(w io.Writer, s Spec) error {
	b, err := glb.Encode(Model(s))
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

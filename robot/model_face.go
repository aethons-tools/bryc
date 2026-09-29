package robot

import (
	"fmt"

	"github.com/aethons-tools/bryc/robot/mesh"
)

// addFace adds the eyes, mouth and jaw to the model.
func addFace(m *modelBuilder) {
	addEyes(m)
	addMouth(m)
}

// place turns a mesh built around the origin into one tilted like the eye
// (canvas clockwise is counter-clockwise once y points up) and moved to the
// canvas point (x, y) at depth z.
func place(me mesh.Mesh, tilt, x, y, z float64) mesh.Mesh {
	return me.RotateZ(-tilt).Translate(upAt(x, y, z))
}

func addEyes(m *modelBuilder) {
	s, b := m.s, m.b
	if s.Eyes == "visor" {
		cx, y := b.CX(), eyeY(s, b)
		poly := roundedRectOutline(cx-visorHalfWidth, y-visorHalfHeight, 2*visorHalfWidth, 2*visorHalfHeight, visorHalfHeight)
		m.add("visor", mustExtrude(upPoly(poly), 40, 8).Translate(mesh.Vec3{Z: faceZ}), m.mat.screen)
		bar := visorHalfWidth - visorBarInset
		bp := roundedRectOutline(cx-bar, y-visorBarHalfHeight, 2*bar, 2*visorBarHalfHeight, visorBarHalfHeight)
		m.add("visor-bar", mustExtrude(upPoly(bp), 10, 3).Translate(mesh.Vec3{Z: faceZ + 22}), m.mat.glow)
		return
	}
	aspect := 1.0
	if s.Eyes != "round" {
		aspect = ovalAspect
	}
	r := roundEyeR(s)
	for i, p := range roundEyeCenters(s, b) {
		x, y := p[0], p[1]
		tilt := eyeTilt(s, b, x)
		name := fmt.Sprintf("eye-%d", i)
		switch s.EyeStyle {
		case "glower":
			k := r / maxEyeRadius
			ring := mesh.Cylinder(r, 0.4*r).Scale(mesh.Vec3{X: 1, Y: aspect, Z: 1})
			m.add(fmt.Sprintf("eye-ring-%d", i), place(ring, tilt, x, y, faceZ), m.mat.metal)
			lr := r - glowerRing*k
			m.add(name, place(mesh.Ellipsoid(lr, lr*aspect, 0.3*r), tilt, x, y, faceZ+0.05*r), m.mat.screen)
			c := glowerCore * k
			m.add(fmt.Sprintf("eye-core-%d", i), place(mesh.Ellipsoid(c, c*aspect, c), tilt, x, y, faceZ+0.35*r), m.mat.glow)
		case "dead":
			m.add(name, place(mesh.Ellipsoid(r, r*aspect, 0.5*r), tilt, x, y, faceZ), m.mat.screen)
		default:
			m.add(name, place(mesh.Ellipsoid(r, r*aspect, 0.5*r), tilt, x, y, faceZ), m.mat.glow)
		}
	}
}

func addMouth(m *modelBuilder) {
	s, b := m.s, m.b
	cx, my := b.CX(), mouthY(s, b)
	flat := func(poly []mesh.Vec2, depth, bevel, z float64) mesh.Mesh {
		return mustExtrude(upPoly(poly), depth, bevel).Translate(mesh.Vec3{Z: z})
	}
	switch s.Mouth {
	case "grille":
		m.add("mouth", flat(bentRectOutline(s.Expression, cx, my, grilleWidth, 2*grilleHalfHeight), 20, 4, faceZ+6), m.mat.metal)
		for k, x := range grilleBars(cx) {
			yc := curve(s.Expression, cx, my, grilleWidth, x)
			poly := []mesh.Vec2{{X: x - 3, Y: yc - (grilleHalfHeight - 4)}, {X: x + 3, Y: yc - (grilleHalfHeight - 4)}, {X: x + 3, Y: yc + (grilleHalfHeight - 4)}, {X: x - 3, Y: yc + (grilleHalfHeight - 4)}}
			m.add(fmt.Sprintf("mouth-bar-%d", k), flat(poly, 4, 1, faceZ+6+10+1), m.mat.ink)
		}
	case "slot":
		m.add("mouth", flat(bentRectOutline(s.Expression, cx, my, slotWidth, slotHeight), 6, 2, faceZ+2), m.mat.screen)
	case "line":
		const steps = 24
		path := make([]mesh.Vec3, 0, steps+1)
		for i := 0; i <= steps; i++ {
			x := cx - lineMouthWidth/2 + lineMouthWidth*float64(i)/steps
			path = append(path, upAt(x, curve(s.Expression, cx, my, lineMouthWidth, x), faceZ+2))
		}
		m.add("mouth", mesh.Tube(path, func(float64) float64 { return 7 }), m.mat.ink)
	case "jaw":
		m.add("jaw", mustExtrude(upPoly(jawOutline(s, b)), 330, 8), m.mat.jaw)
		for i, p := range jawBolts(s, b) {
			m.add(fmt.Sprintf("jaw-bolt-%d", i), mesh.Ellipsoid(jawBoltRadius, jawBoltRadius, 5).Translate(upAt(p[0], p[1], 165)), m.mat.metal)
		}
	}
}

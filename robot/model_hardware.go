package robot

import (
	"fmt"
	"math"

	"github.com/aethons-tools/bryc/robot/mesh"
)

// addHardware adds the rivets, seams, blush, eyelashes, ears and antenna to
// the model.
const (
	// seamInset keeps seam ends clear of the head's bevelled edge: the
	// bevel plus the seam's radius.
	seamInset = headBevel + 4
	// dialSink is how far a dial's disc reaches into the head, so it stays
	// attached on slanted or curved sides; dialDisc is the disc's length and
	// capLen the cap's.
	dialSink, dialDisc, capLen = 15.0, 40.0, 10.0
)

func addHardware(m *modelBuilder) {
	addFaceDetails(m)
	addEars(m)
	addAntenna(m)
}

func addFaceDetails(m *modelBuilder) {
	s, b := m.s, m.b
	if s.Rivets != nil && *s.Rivets {
		for i, p := range rivetCenters(s, b) {
			m.add(fmt.Sprintf("rivet-%d", i), mesh.Ellipsoid(rivetRadius, rivetRadius, 6).Translate(upAt(p[0], p[1], faceZ)), m.mat.metal)
		}
	}
	if s.Panels != nil && *s.Panels {
		const steps = 16
		for i, seam := range panelSeams(s, b) {
			x0 := leftEdge(s.Head, b, seam.y)
			x1 := 2*b.CX() - x0
			sag := math.Abs(x1-x0) * bowRatio
			path := make([]mesh.Vec3, 0, steps+1)
			// The bow is sized to the full edge-to-edge width (as in 2D), but
			// only the span inside the bevel is sampled.
			t0 := seamInset / (x1 - x0)
			for k := 0; k <= steps; k++ {
				t := t0 + (1-2*t0)*float64(k)/steps
				y := seam.y
				if headBows[s.Head] {
					// The 2D bowTo quadratic: its control point is 2*sag off.
					y += seam.dir * 2 * sag * 2 * t * (1 - t)
				}
				path = append(path, upAt(x0+(x1-x0)*t, y, faceZ+1))
			}
			m.add(fmt.Sprintf("seam-%d", i), mesh.Tube(path, func(float64) float64 { return 4 }), m.mat.ink)
		}
	}
	if s.Blush != nil && *s.Blush {
		// The 2D blush is drawn over the jaw, so with a jaw it sits on the
		// jaw's front rather than buried inside it.
		z := faceZ + 2
		if s.Mouth == "jaw" {
			z = jawFront + 2
		}
		for i, p := range blushCenters(s, b) {
			poly := ellipseOutline(p[0], p[1], blushRX, blushRY)
			m.add(fmt.Sprintf("blush-%d", i), mustExtrude(upPoly(poly), 2, 0.5).Translate(mesh.Vec3{Z: z}), m.mat.blush)
		}
	}
	if s.Eyes != "visor" && hasLashes(s) {
		const samples = 21
		r := roundEyeR(s)
		for i, p := range roundEyeCenters(s, b) {
			l := lashFor(s, b, p[0], p[1], r)
			path := make([]mesh.Vec3, 0, samples)
			for k := 0; k < samples; k++ {
				x, y := l.at(float64(k) / (samples - 1))
				path = append(path, upAt(x, y, faceZ+2))
			}
			m.add(fmt.Sprintf("lash-%d", i), mesh.Tube(path, func(t float64) float64 { return max(l.halfWidth(t), 0.5) }), m.mat.ink)
		}
	}
}

func addEars(m *modelBuilder) {
	s, b := m.s, m.b
	switch s.Ears {
	case "none":
	case "bolts":
		blocks, dots := boltEars(s, b)
		for i, r := range blocks {
			poly := roundedRectOutline(r.x, r.y, r.w, r.h, r.r)
			m.add(fmt.Sprintf("ear-%d", i), mustExtrude(upPoly(poly), 240, 10), m.mat.accent)
		}
		for i, d := range dots {
			m.add(fmt.Sprintf("ear-dot-%d", i), mesh.Ellipsoid(12, 12, 6).Translate(upAt(d.x, d.y, 120)), m.mat.ink)
		}
	case "dials":
		for i, d := range dialEars(s, b) {
			// out is +1 on the right, -1 on the left; the disc's inner face
			// sits on the head's side and everything else stacks outward.
			out := 1.0
			if i == 0 {
				out = -1
			}
			edge := leftEdge(s.Head, b, d.y)
			if i == 1 {
				edge = 2*b.CX() - edge
			}
			disc := mesh.Cylinder(dialRadius, dialDisc).RotateY(math.Pi / 2)
			m.add(fmt.Sprintf("ear-%d", i), disc.Translate(upAt(edge+out*(dialDisc/2-dialSink), d.y, 0)), m.mat.accent)
			cap := mesh.Cylinder(dialCapRadius, capLen).RotateY(math.Pi / 2)
			m.add(fmt.Sprintf("ear-cap-%d", i), cap.Translate(upAt(edge+out*(dialDisc-dialSink+capLen/2), d.y, 0)), m.mat.body)
			tick := []mesh.Vec3{upAt(edge+out*(dialDisc-dialSink+capLen), d.y, 0), upAt(edge+out*(dialDisc-dialSink+capLen), d.y-dialTick, 0)}
			m.add(fmt.Sprintf("ear-tick-%d", i), mesh.Tube(tick, func(float64) float64 { return 4 }), m.mat.ink)
		}
	default:
		panic("robot: unknown ears " + s.Ears)
	}
}

func addAntenna(m *modelBuilder) {
	a := antennaGeometry(m.s, m.b)
	if len(a.bolt) > 0 {
		path := make([]mesh.Vec3, len(a.bolt))
		for i, p := range a.bolt {
			path[i] = upAt(p[0], p[1], 0)
		}
		m.add("antenna-bolt", mesh.Tube(path, func(float64) float64 { return 9 }), m.mat.accent)
	}
	for k, st := range a.stems {
		path := []mesh.Vec3{upAt(st[0][0], st[0][1], 0), upAt(st[1][0], st[1][1], 0)}
		m.add(fmt.Sprintf("antenna-stem-%d", k), mesh.Tube(path, func(float64) float64 { return 7 }), m.mat.metal)
	}
	for k, ball := range a.balls {
		m.add(fmt.Sprintf("antenna-ball-%d", k), mesh.Ellipsoid(ball.r, ball.r, ball.r).Translate(upAt(ball.x, ball.y, 0)), m.mat.accent)
	}
	for k, base := range a.bases {
		poly := roundedRectOutline(base.x, base.y, base.w, base.h, base.r)
		m.add(fmt.Sprintf("antenna-base-%d", k), mustExtrude(upPoly(poly), 90, 8), m.mat.antennaBase)
	}
}

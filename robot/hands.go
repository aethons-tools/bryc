package robot

import (
	"math"

	"github.com/aethons-tools/bryc/robot/mesh"
)

// Hands float in front of the robot below its chin, palms down as if on a
// keyboard: we see the backs of the hands, fingers pointing down (toward
// us), thumbs toward the middle. 2D and 3D share this geometry.

// hand is one hand. Its shape is given as offsets from the wrist in "hand
// space": u across the hand, v along it toward the fingertips. 2D draws
// hand space straight onto the canvas (v down); the 3D model lays it nearly
// flat in front of the face.
type hand struct {
	wrist  mesh.Vec2   // canvas position of the wrist
	palm   []mesh.Vec2 // palm outline, offsets from the wrist
	digits [3]digit    // the two fingers, then the thumb
}

// digit is a finger or thumb: a capsule from its root (inside the palm) to
// the centre of its round tip.
type digit struct {
	root, tip mesh.Vec2
	r         float64
}

// Hand placement and shape, in virtual units, for the left hand (on the
// viewer's left, so its thumb points right, toward the middle). The right
// hand is its mirror image.
const (
	handSpread = 170.0 // wrist distance from the head's centre line
	handWristY = 712.0 // wrist height: below the lowest mouth, overlapping the chin
	handTilt   = 18.0  // degrees the fingers turn in toward the middle
	palmRound  = 30.0  // corner radius of the palm's triangle
)

var (
	// palmCorners is the palm's triangle before rounding: a point at the
	// wrist and the knuckle edge below it.
	palmCorners = []mesh.Vec2{{X: 0, Y: 20}, {X: 40, Y: 78}, {X: -40, Y: 78}}
	leftDigits  = [3]digit{
		{root: mesh.Vec2{X: -24, Y: 74}, tip: mesh.Vec2{X: -42, Y: 140}, r: 19},
		{root: mesh.Vec2{X: 14, Y: 78}, tip: mesh.Vec2{X: 16, Y: 146}, r: 19},
		{root: mesh.Vec2{X: 30, Y: 42}, tip: mesh.Vec2{X: 86, Y: 70}, r: 16},
	}
)

// hands are the robot's two hands, left first, for the head in b.
func hands(b Layout) [2]hand {
	a := handTilt * math.Pi / 180
	sin, cos := math.Sin(a), math.Cos(a)
	var out [2]hand
	for i, side := range []float64{-1, 1} {
		// Tilt the fingers in toward the middle, then mirror for the right.
		place := func(p mesh.Vec2) mesh.Vec2 {
			u, v := p.X*cos+p.Y*sin, -p.X*sin+p.Y*cos
			return mesh.Vec2{X: -side * u, Y: v}
		}
		h := hand{wrist: mesh.Vec2{X: b.CX() + side*handSpread, Y: handWristY}}
		for _, p := range roundedConvexOutline(palmCorners, palmRound) {
			h.palm = append(h.palm, place(p))
		}
		for j, d := range leftDigits {
			h.digits[j] = digit{root: place(d.root), tip: place(d.tip), r: d.r}
		}
		out[i] = h
	}
	return out
}

// at is the canvas position of an offset from the wrist.
func (h hand) at(p mesh.Vec2) mesh.Vec2 { return mesh.Vec2{X: h.wrist.X + p.X, Y: h.wrist.Y + p.Y} }

// roundedConvexOutline is the outline of circles of radius r centred on
// the corners of the convex polygon pts, joined by their outer tangents:
// the polygon grown by r, with round corners.
func roundedConvexOutline(pts []mesh.Vec2, r float64) []mesh.Vec2 {
	n := len(pts)
	var cx, cy float64
	for _, p := range pts {
		cx, cy = cx+p.X/float64(n), cy+p.Y/float64(n)
	}
	// outward is the outward normal angle of edge a→b.
	outward := func(a, b mesh.Vec2) float64 {
		nx, ny := b.Y-a.Y, a.X-b.X
		if nx*((a.X+b.X)/2-cx)+ny*((a.Y+b.Y)/2-cy) < 0 {
			nx, ny = -nx, -ny
		}
		return math.Atan2(ny, nx)
	}
	var out []mesh.Vec2
	for i, p := range pts {
		a0 := outward(pts[(i+n-1)%n], p)
		a1 := outward(p, pts[(i+1)%n])
		d := math.Remainder(a1-a0, 2*math.Pi) // the exterior angle, signed by winding
		for k := 0; k <= quarterSteps; k++ {
			a := a0 + d*float64(k)/quarterSteps
			out = append(out, mesh.Vec2{X: p.X + r*math.Cos(a), Y: p.Y + r*math.Sin(a)})
		}
	}
	return out
}

// capsuleOutline is the outline of the capsule of radius r from a to b.
func capsuleOutline(a, b mesh.Vec2, r float64) []mesh.Vec2 {
	dir := math.Atan2(b.Y-a.Y, b.X-a.X)
	var out []mesh.Vec2
	for _, end := range []struct {
		c     mesh.Vec2
		start float64
	}{{b, dir - math.Pi/2}, {a, dir + math.Pi/2}} {
		for k := 0; k <= 2*quarterSteps; k++ {
			t := end.start + math.Pi*float64(k)/(2*quarterSteps)
			out = append(out, mesh.Vec2{X: end.c.X + r*math.Cos(t), Y: end.c.Y + r*math.Sin(t)})
		}
	}
	return out
}

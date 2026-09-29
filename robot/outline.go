package robot

import (
	"math"

	"github.com/aethons-tools/bryc/robot/mesh"
)

// Outline samplers: the 2D shapes as explicit polygons in canvas coordinates
// (y down), for building 3D meshes that line up with the 2D drawing. Every
// sampler uses fixed step counts so its output is deterministic. Outlines are
// open lists: the closing edge back to the first point is implied.

// Sampling densities.
const (
	quarterSteps = 8  // segments per quarter circle
	halfSteps    = 32 // segments for a dome's half circle
	bowSteps     = 16 // segments for a bowed edge (see bowTo)
	ellipseSteps = 32 // points around an ellipse
)

// pathBuilder collects outline points, skipping repeats of the last point.
type pathBuilder []mesh.Vec2

func samePoint(a, b mesh.Vec2) bool {
	return math.Abs(a.X-b.X) < 1e-9 && math.Abs(a.Y-b.Y) < 1e-9
}

func (p *pathBuilder) add(x, y float64) {
	v := mesh.Vec2{X: x, Y: y}
	if n := len(*p); n > 0 && samePoint((*p)[n-1], v) {
		return
	}
	*p = append(*p, v)
}

// arc adds the arc of radius r about (cx, cy) from angle a0 to a1 in steps
// segments, with gg's y-down convention (angles increase clockwise on screen).
func (p *pathBuilder) arc(cx, cy, r, a0, a1 float64, steps int) {
	for i := 0; i <= steps; i++ {
		a := a0 + (a1-a0)*float64(i)/float64(steps)
		p.add(cx+r*math.Cos(a), cy+r*math.Sin(a))
	}
}

// bow adds the bowed edge bowTo traces from (from, y) to (to, y).
func (p *pathBuilder) bow(from, to, y, dir float64) {
	sag := math.Abs(to-from) * bowRatio
	cx, cy := (from+to)/2, y+dir*2*sag
	for i := 0; i <= bowSteps; i++ {
		t := float64(i) / bowSteps
		u := 1 - t
		p.add(u*u*from+2*u*t*cx+t*t*to, u*u*y+2*u*t*cy+t*t*y)
	}
}

// done returns the outline, dropping a final point that repeats the first.
func (p pathBuilder) done() []mesh.Vec2 {
	if n := len(p); n > 1 && samePoint(p[0], p[n-1]) {
		p = p[:n-1]
	}
	return p
}

// roundedRectOutline is the rounded rectangle gg's DrawRoundedRectangle
// draws, clockwise on screen from the top-left corner.
func roundedRectOutline(x, y, w, h, r float64) []mesh.Vec2 {
	var p pathBuilder
	if r <= 0 {
		p.add(x, y)
		p.add(x+w, y)
		p.add(x+w, y+h)
		p.add(x, y+h)
		return p.done()
	}
	p.arc(x+r, y+r, r, math.Pi, 1.5*math.Pi, quarterSteps)
	p.arc(x+w-r, y+r, r, 1.5*math.Pi, 2*math.Pi, quarterSteps)
	p.arc(x+w-r, y+h-r, r, 0, 0.5*math.Pi, quarterSteps)
	p.arc(x+r, y+h-r, r, 0.5*math.Pi, math.Pi, quarterSteps)
	return p.done()
}

// ellipseOutline is the ellipse with radii rx, ry about (cx, cy).
func ellipseOutline(cx, cy, rx, ry float64) []mesh.Vec2 {
	p := make([]mesh.Vec2, ellipseSteps)
	for i := range p {
		a := 2 * math.Pi * float64(i) / ellipseSteps
		p[i] = mesh.Vec2{X: cx + rx*math.Cos(a), Y: cy + ry*math.Sin(a)}
	}
	return p
}

// headOutline is s's head outline in b, following headPaths.
func headOutline(s Spec, b Layout) []mesh.Vec2 {
	var p pathBuilder
	X, Y, W, B := b.X, b.Y, b.W, b.Y+b.H
	switch s.Head {
	case "square", "rounded":
		return roundedRectOutline(b.X, b.Y, b.W, b.H, cornerRadius[s.Head])
	case "dome":
		r := W / 2
		p.add(X, B)
		p.add(X, Y+r)
		p.arc(b.CX(), Y+r, r, math.Pi, 2*math.Pi, halfSteps)
		p.add(X+W, B)
		p.bow(X+W, X, B, +1)
	case "inverted-dome":
		r := W / 2
		p.add(X, Y)
		p.bow(X, X+W, Y, -1)
		p.add(X+W, B-r)
		p.arc(b.CX(), B-r, r, 0, math.Pi, halfSteps)
	case "trapezoid":
		p.add(X+trapezoidInset, Y)
		p.bow(X+trapezoidInset, X+W-trapezoidInset, Y, -1)
		p.add(X+W, B)
		p.bow(X+W, X, B, +1)
	case "inverted-trapezoid":
		p.add(X, Y)
		p.bow(X, X+W, Y, -1)
		p.add(X+W-trapezoidInset, B)
		p.bow(X+W-trapezoidInset, X+trapezoidInset, B, +1)
	}
	return p.done()
}

// minShoulderRound is the smallest shoulder corner radius the outline
// rounds: the 3D model's bevel. Smaller corners are traced sharp, since a
// bevel can't follow a curve tighter than itself.
const minShoulderRound = 12.0

// minShoulderEdge is the shortest edge the shoulder outline keeps between
// sampled corner points, so the 3D bevel never meets a sliver of an edge.
const minShoulderEdge = 18.0

// shoulderOutline is the shoulders' outline (see shoulderPath), cut off at
// the bottom of the canvas. It is simplified for the 3D bevel: tiny corner
// radii are traced sharp, and corner samples that crowd their neighbours
// (at the bottom cut or where an arc meets a short edge) are dropped.
func shoulderOutline(shoulders int) []mesh.Vec2 {
	var p pathBuilder
	t := float64(shoulders) / ShouldersMax
	p.add(shoulderLeft, shoulderBottom)
	if t >= 0 {
		r := maxShoulderRadius * t
		if r < minShoulderRound {
			r = 0
		}
		p.add(shoulderLeft, shoulderTop+r)
		if r > 0 {
			p.arc(shoulderLeft+r, shoulderTop+r, r, math.Pi, 1.5*math.Pi, quarterSteps)
		}
		p.add(shoulderRight-r, shoulderTop)
		if r > 0 {
			p.arc(shoulderRight-r, shoulderTop+r, r, 1.5*math.Pi, 2*math.Pi, quarterSteps)
		}
	} else {
		h, lean := maxSpikeHeight*-t, maxSpikeLean*-t
		p.add(shoulderLeft, shoulderTop)
		p.add(shoulderLeft-lean, shoulderTop-h)
		p.add(shoulderLeft+spikeBase, shoulderTop)
		p.add(shoulderRight-spikeBase, shoulderTop)
		p.add(shoulderRight+lean, shoulderTop-h)
		p.add(shoulderRight, shoulderTop)
	}
	p.add(shoulderRight, shoulderBottom)
	// Large corner radii reach below the canvas, so trace the full shape
	// and clip it rather than just moving the bottom edge up.
	cut := clipY(p.done(), virtual, false)
	if t <= 0 {
		return cut // no arcs: every point is a corner or spike tip
	}
	// Only arc samples (and the straight edges' ends, which are arc samples
	// too) may go; the points on the cut stay.
	return thin(cut, minShoulderEdge, func(v mesh.Vec2) bool { return v.Y != virtual })
}

// thin drops droppable points that lie closer than gap to a kept neighbour:
// walking the outline, a droppable point is skipped if it is too close to
// the last kept point, and a fixed point removes any droppable points kept
// just before it that are too close to it.
func thin(poly []mesh.Vec2, gap float64, droppable func(mesh.Vec2) bool) []mesh.Vec2 {
	near := func(a, b mesh.Vec2) bool { return math.Hypot(a.X-b.X, a.Y-b.Y) < gap }
	var out []mesh.Vec2
	for _, v := range poly {
		if droppable(v) {
			if len(out) > 0 && near(out[len(out)-1], v) {
				continue
			}
		} else {
			for len(out) > 0 && droppable(out[len(out)-1]) && near(out[len(out)-1], v) {
				out = out[:len(out)-1]
			}
		}
		out = append(out, v)
	}
	// The closing edge, back to the first point.
	for len(out) > 1 && droppable(out[len(out)-1]) && near(out[len(out)-1], out[0]) {
		out = out[:len(out)-1]
	}
	return out
}

// clipY clips poly (Sutherland–Hodgman) to one side of the line y = at:
// keepBelow keeps y >= at (below it on screen), otherwise y <= at. Points
// created on the cut have Y exactly at.
func clipY(poly []mesh.Vec2, at float64, keepBelow bool) []mesh.Vec2 {
	inside := func(v mesh.Vec2) bool {
		if keepBelow {
			return v.Y >= at
		}
		return v.Y <= at
	}
	var p pathBuilder
	for i := range poly {
		a, b := poly[i], poly[(i+1)%len(poly)]
		if inside(a) {
			p.add(a.X, a.Y)
		}
		if inside(a) != inside(b) {
			p.add(a.X+(at-a.Y)*(b.X-a.X)/(b.Y-a.Y), at)
		}
	}
	return p.done()
}

// bentRectOutline is the outline bentRect traces: 25 points along the top
// edge, left to right, then 25 along the bottom, right to left.
func bentRectOutline(expression string, cx, y, w, h float64) []mesh.Vec2 {
	const steps = 24
	p := make([]mesh.Vec2, 0, 2*(steps+1))
	for i := 0; i <= steps; i++ {
		x := cx - w/2 + w*float64(i)/steps
		p = append(p, mesh.Vec2{X: x, Y: curve(expression, cx, y, w, x) - h/2})
	}
	for i := steps; i >= 0; i-- {
		x := cx - w/2 + w*float64(i)/steps
		p = append(p, mesh.Vec2{X: x, Y: curve(expression, cx, y, w, x) + h/2})
	}
	return p
}

// jawTop is the height of the jaw's flat arm tops.
func jawTop(s Spec, b Layout) float64 { return mouthY(s, b) - b.H*jawHeight }

// jawFrame is the jaw geometry drawJaw and the jaw outline share: the mouth
// line, the arms' top, the scaling center, the inner edge of the left arm at
// height y, and the outer edge of the left arm at its top.
type jawFrame struct {
	my, top, cx, cy, outerL float64
	innerL                  func(y float64) float64
}

func jawFrameFor(s Spec, b Layout) jawFrame {
	my := mouthY(s, b)
	top := my - b.H*jawHeight
	cx, cy := b.CX(), b.Y+b.H/2
	innerL := func(y float64) float64 { return leftEdge(s.Head, b, y) + jawBand }
	outerL := cx - jawOverhang*(cx-leftEdge(s.Head, b, cy+(top-cy)/jawOverhang))
	return jawFrame{my: my, top: top, cx: cx, cy: cy, outerL: outerL, innerL: innerL}
}

// jawBoltRadius is the radius of the hinge bolt on each arm of the jaw.
const jawBoltRadius = 9.0

// jawBolts are the centers of the hinge bolts on the jaw's arms, left first.
func jawBolts(s Spec, b Layout) [2][2]float64 {
	j := jawFrameFor(s, b)
	var bolts [2][2]float64
	for i, side := range []float64{-1, 1} {
		bolts[i] = [2]float64{j.cx + side*(j.cx-(j.outerL+j.innerL(j.top+28))/2), j.top + 28}
	}
	return bolts
}

// jawInner is the jaw's inner edge, following drawJaw's cut-out: down the
// left arm's inside from the top, along the mouth line, and up the right arm.
func jawInner(s Spec, b Layout) []mesh.Vec2 {
	j := jawFrameFor(s, b)
	const steps = 24
	x0 := j.innerL(j.my)
	w := 2 * (j.cx - x0)
	endY := curve(s.Expression, j.cx, j.my, w, x0)
	// A frown's ends drop below the mouth line, and on heads that narrow
	// toward the chin the arms' inner edges there pass the mouth's ends (the
	// 2D even-odd fill hides the resulting sliver), so the sides stop at the
	// mouth line to keep the outline simple.
	// The right arm mirrors the left arm's samples exactly (drawJaw steps
	// the right arm up from the bottom instead), so the plate is equally
	// thick on both sides.
	sideEnd := min(endY, j.my)
	var left []mesh.Vec2
	for y := j.top; y < sideEnd; y += 10 {
		left = append(left, mesh.Vec2{X: j.innerL(y), Y: y})
	}
	var p pathBuilder
	for _, v := range left {
		p.add(v.X, v.Y)
	}
	for i := 0; i <= steps; i++ {
		x := x0 + w*float64(i)/steps
		p.add(x, curve(s.Expression, j.cx, j.my, w, x))
	}
	for i := len(left) - 1; i >= 0; i-- {
		p.add(2*j.cx-left[i].X, left[i].Y)
	}
	return p
}

// jawOutline is the jaw plate drawJaw fills, as one simple U-shaped polygon:
// the head outline scaled by jawOverhang, cut at the jaw's top, with its
// top edge replaced by the inner edge of the U.
func jawOutline(s Spec, b Layout) []mesh.Vec2 {
	top := jawTop(s, b)
	cx, cy := b.CX(), b.Y+b.H/2
	head := headOutline(s, b)
	outer := make([]mesh.Vec2, len(head))
	for i, v := range head {
		outer[i] = mesh.Vec2{X: cx + (v.X-cx)*jawOverhang, Y: cy + (v.Y-cy)*jawOverhang}
	}
	outer = clipY(outer, top, true)
	inner := jawInner(s, b)
	for i := range outer {
		a, c := outer[i], outer[(i+1)%len(outer)]
		if a.Y != top || c.Y != top {
			continue
		}
		// The inner path runs left to right; follow the cut edge's direction.
		if a.X > c.X {
			for l, r := 0, len(inner)-1; l < r; l, r = l+1, r-1 {
				inner[l], inner[r] = inner[r], inner[l]
			}
		}
		var p pathBuilder
		for _, v := range outer[:i+1] {
			p.add(v.X, v.Y)
		}
		for _, v := range inner {
			p.add(v.X, v.Y)
		}
		for _, v := range outer[i+1:] {
			p.add(v.X, v.Y)
		}
		return p.done()
	}
	return outer // the jaw's top is above the head: no arms to cut between
}

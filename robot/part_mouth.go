package robot

import "math"

func mouthY(s Spec, b Layout) float64 { _, my := facePos(s, b); return my }

// maxBend is how far, in virtual units, the ends of a mouth at least
// bendWidth wide move up (smile) or down (frown). Narrower mouths bend
// proportionally less so they stay "slight".
const (
	maxBend   = 22.0
	bendWidth = 240.0
)

// Jaw geometry: the jaw is a U-shaped plate hugging the lower face.
const (
	jawShade    = 0.82 // jaw color relative to the body color
	jawBand     = 42.0 // width of the U's arms up the cheeks
	jawOverhang = 1.04 // scale of the head outline the jaw follows, so it overlaps the edge
	jawHeight   = 0.18 // how far the arms rise above the mouth line, as a fraction of head height
)

// curve returns a mouth's centerline height at x for a mouth centered at cx
// and y with the given width: flat, or a shallow parabola whose ends rise for
// a smile and drop for a frown.
func curve(expression string, cx, y, width, x float64) float64 {
	bend := maxBend * math.Min(1, width/bendWidth)
	u := (x - cx) / (width / 2)
	switch expression {
	case "smile":
		return y - bend*u*u
	case "frown":
		return y + bend*u*u
	}
	return y
}

// bentRect traces a rectangle of the given size whose top and bottom edges
// follow the expression's curve.
func bentRect(c *canvas, expression string, cx, y, w, h float64) {
	const steps = 24
	c.dc.NewSubPath()
	for i := 0; i <= steps; i++ {
		x := cx - w/2 + w*float64(i)/steps
		c.dc.LineTo(x, curve(expression, cx, y, w, x)-h/2)
	}
	for i := steps; i >= 0; i-- {
		x := cx - w/2 + w*float64(i)/steps
		c.dc.LineTo(x, curve(expression, cx, y, w, x)+h/2)
	}
	c.dc.ClosePath()
}

var mouthParts = map[string]part{
	"grille": func(c *canvas, s Spec, b Layout) {
		const w, h = 240.0, 70.0
		cx, y := b.CX(), mouthY(s, b)
		bentRect(c, s.Expression, cx, y, w, h)
		c.fillOutlined(metal)
		for x := cx - 80; x <= cx+80; x += 40 {
			yc := curve(s.Expression, cx, y, w, x)
			c.dc.MoveTo(x, yc-h/2)
			c.dc.LineTo(x, yc+h/2)
		}
		c.strokeWith(ink, 8)
	},
	"slot": func(c *canvas, s Spec, b Layout) {
		bentRect(c, s.Expression, b.CX(), mouthY(s, b), 120, 36)
		c.fillOutlined(screen)
	},
	"line": func(c *canvas, s Spec, b Layout) {
		const w, steps = 180.0, 24
		cx, y := b.CX(), mouthY(s, b)
		for i := 0; i <= steps; i++ {
			x := cx - w/2 + w*float64(i)/steps
			c.dc.LineTo(x, curve(s.Expression, cx, y, w, x))
		}
		c.strokeWith(ink, 14)
	},
	"jaw": drawJaw,
}

// drawJaw draws a U-shaped plate that overlaps the lower face and hugs its
// outline: arms up both cheeks and a chin whose top edge is the mouth.
//
// The plate is the head outline scaled up slightly (so it overhangs the
// edge), clipped to below the jaw's top, with the face above the mouth line
// cut out using the even-odd fill rule.
func drawJaw(c *canvas, s Spec, b Layout) {
	dc := c.dc
	my := mouthY(s, b)
	top := my - b.H*jawHeight
	cx, cy := b.CX(), b.Y+b.H/2
	innerL := func(y float64) float64 { return leftEdge(s.Head, b, y) + jawBand }
	outerL := cx - jawOverhang*(cx-leftEdge(s.Head, b, cy+(top-cy)/jawOverhang))

	// Clip to everything below the jaw's top edge.
	dc.DrawRectangle(0, top, virtual, virtual-top)
	dc.Clip()

	// Outer outline: the head, scaled about its center.
	dc.Push()
	dc.ScaleAbout(jawOverhang, jawOverhang, cx, cy)
	headPaths[s.Head](dc, b)
	dc.Pop()

	// Cut-out: the face between the arms, down to the mouth line. It starts
	// above the clip so only its sides and the mouth line get outlined.
	const steps = 24
	x0 := innerL(my)
	w := 2 * (cx - x0)
	endY := curve(s.Expression, cx, my, w, x0) // where the arms meet the mouth line
	dc.NewSubPath()
	dc.MoveTo(innerL(top-40), top-40)
	for y := top; y < endY; y += 10 {
		dc.LineTo(innerL(y), y)
	}
	for i := 0; i <= steps; i++ {
		x := x0 + w*float64(i)/steps
		dc.LineTo(x, curve(s.Expression, cx, my, w, x))
	}
	for y := endY - 10; y >= top; y -= 10 {
		dc.LineTo(2*cx-innerL(y), y)
	}
	dc.LineTo(2*cx-innerL(top-40), top-40)
	dc.ClosePath()

	dc.SetFillRuleEvenOdd()
	c.fillOutlined(shade(parseHex(s.Body), jawShade))
	dc.SetFillRuleWinding()
	dc.ResetClip()

	// Outline the flat tops of the arms, which the clip left open, and add a
	// hinge bolt on each.
	for _, side := range []float64{-1, 1} {
		dc.MoveTo(cx+side*(cx-outerL), top)
		dc.LineTo(cx+side*(cx-innerL(top)), top)
		c.strokeWith(ink, outline)
		dc.DrawCircle(cx+side*(cx-(outerL+innerL(top+28))/2), top+28, 9)
		dc.SetColor(metal)
		dc.FillPreserve()
		dc.SetColor(ink)
		c.lw(5)
		dc.Stroke()
	}
}

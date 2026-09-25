package robot

import (
	"math"

	"github.com/fogleman/gg"
)

// headPaths trace each head outline as the current path without drawing it,
// so the outline can be filled, or reused as a clip (see drawPanels).
var headPaths = map[string]func(dc *gg.Context, b Layout){
	"square": func(dc *gg.Context, b Layout) {
		dc.DrawRoundedRectangle(b.X, b.Y, b.W, b.H, cornerRadius["square"])
	},
	"rounded": func(dc *gg.Context, b Layout) {
		dc.DrawRoundedRectangle(b.X, b.Y, b.W, b.H, cornerRadius["rounded"])
	},
	"dome": func(dc *gg.Context, b Layout) {
		r := b.W / 2
		dc.NewSubPath()
		dc.MoveTo(b.X, b.Y+b.H)
		dc.LineTo(b.X, b.Y+r)
		dc.DrawArc(b.CX(), b.Y+r, r, math.Pi, 2*math.Pi)
		dc.LineTo(b.X+b.W, b.Y+b.H)
		dc.ClosePath()
	},
	"inverted-dome": func(dc *gg.Context, b Layout) {
		r := b.W / 2
		dc.NewSubPath()
		dc.MoveTo(b.X, b.Y)
		dc.LineTo(b.X+b.W, b.Y)
		dc.LineTo(b.X+b.W, b.Y+b.H-r)
		dc.DrawArc(b.CX(), b.Y+b.H-r, r, 0, math.Pi)
		dc.ClosePath()
	},
	"trapezoid": func(dc *gg.Context, b Layout) {
		dc.NewSubPath()
		dc.MoveTo(b.X+trapezoidInset, b.Y)
		dc.LineTo(b.X+b.W-trapezoidInset, b.Y)
		dc.LineTo(b.X+b.W, b.Y+b.H)
		dc.LineTo(b.X, b.Y+b.H)
		dc.ClosePath()
	},
	"inverted-trapezoid": func(dc *gg.Context, b Layout) {
		dc.NewSubPath()
		dc.MoveTo(b.X, b.Y)
		dc.LineTo(b.X+b.W, b.Y)
		dc.LineTo(b.X+b.W-trapezoidInset, b.Y+b.H)
		dc.LineTo(b.X+trapezoidInset, b.Y+b.H)
		dc.ClosePath()
	},
}

// cornerRadius is the bottom-corner radius of each head shape (0 if sharp).
var cornerRadius = map[string]float64{"square": 18, "rounded": 120}

// bottomRadius is the radius of the head's bottom corners; the inverted
// dome's bottom is a half circle, so its corners are half the head's width.
func bottomRadius(head string, b Layout) float64 {
	if head == "inverted-dome" {
		return b.W / 2
	}
	return cornerRadius[head]
}

// trapezoidInset is how far the trapezoid head's narrow end (its top, or its
// bottom when inverted) sits inside the box.
const trapezoidInset = 60

// sideInset is how far the head's side edge sits inside the box at height y,
// so side-mounted parts such as ears stay attached to slanted heads.
func sideInset(head string, b Layout, y float64) float64 {
	switch head {
	case "trapezoid":
		return trapezoidInset * (1 - (y-b.Y)/b.H)
	case "inverted-trapezoid":
		return trapezoidInset * (y - b.Y) / b.H
	}
	return 0
}

// leftEdge is the x of the head's left outline at height y, including the
// dome's cap and rounded bottom corners (the top corners of square and
// rounded heads are ignored). The right edge mirrors it about b.CX().
func leftEdge(head string, b Layout, y float64) float64 {
	x := b.X + sideInset(head, b, y)
	if r := b.W / 2; head == "dome" && y < b.Y+r {
		dy := b.Y + r - y
		return b.CX() - math.Sqrt(max(0, r*r-dy*dy))
	}
	r := bottomRadius(head, b)
	if dy := y - (b.Y + b.H - r); r > 0 && dy > 0 {
		x += r - math.Sqrt(max(0, r*r-dy*dy))
	}
	return x
}

func drawHead(c *canvas, s Spec, b Layout) {
	headPaths[s.Head](c.dc, b)
	c.fillOutlined(parseHex(s.Body))
}

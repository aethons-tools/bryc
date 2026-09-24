package robot

import (
	"math"

	"github.com/fogleman/gg"
)

// headPaths trace each head outline as the current path without drawing it,
// so the outline can be filled, or reused as a clip (see drawPanels).
var headPaths = map[string]func(dc *gg.Context, b Layout){
	"square": func(dc *gg.Context, b Layout) {
		dc.DrawRoundedRectangle(b.X, b.Y, b.W, b.H, 18)
	},
	"rounded": func(dc *gg.Context, b Layout) {
		dc.DrawRoundedRectangle(b.X, b.Y, b.W, b.H, 120)
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
	"trapezoid": func(dc *gg.Context, b Layout) {
		dc.NewSubPath()
		dc.MoveTo(b.X+trapezoidInset, b.Y)
		dc.LineTo(b.X+b.W-trapezoidInset, b.Y)
		dc.LineTo(b.X+b.W, b.Y+b.H)
		dc.LineTo(b.X, b.Y+b.H)
		dc.ClosePath()
	},
}

// trapezoidInset is how far the trapezoid head's top corners sit inside the box.
const trapezoidInset = 60

// sideInset is how far the head's side edge sits inside the box at height y,
// so side-mounted parts such as ears stay attached to slanted heads.
func sideInset(head string, b Layout, y float64) float64 {
	if head != "trapezoid" {
		return 0
	}
	return trapezoidInset * (1 - (y-b.Y)/b.H)
}

func drawHead(c *canvas, s Spec, b Layout) {
	headPaths[s.Head](c.dc, b)
	c.fillOutlined(parseHex(s.Body))
}

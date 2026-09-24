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
		const inset = 60
		dc.NewSubPath()
		dc.MoveTo(b.X+inset, b.Y)
		dc.LineTo(b.X+b.W-inset, b.Y)
		dc.LineTo(b.X+b.W, b.Y+b.H)
		dc.LineTo(b.X, b.Y+b.H)
		dc.ClosePath()
	},
}

func drawHead(c *canvas, s Spec, b Layout) {
	headPaths[s.Head](c.dc, b)
	c.fillOutlined(parseHex(s.Body))
}

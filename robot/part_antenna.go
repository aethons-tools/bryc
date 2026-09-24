package robot

// antennaBase draws the little mount where an antenna meets the head.
func antennaBase(c *canvas, s Spec, x, y float64) {
	c.dc.DrawRoundedRectangle(x-45, y-22, 90, 34, 10)
	c.fillOutlined(shade(parseHex(s.Body), 0.8))
}

func stem(c *canvas, x1, y1, x2, y2 float64) {
	c.dc.MoveTo(x1, y1)
	c.dc.LineTo(x2, y2)
	c.strokeWith(ink, 14)
}

var antennaParts = map[string]part{
	"none": func(*canvas, Spec, Layout) {},
	"ball": func(c *canvas, s Spec, b Layout) {
		x, top := b.CX(), b.Y
		stem(c, x, top, x, top-110)
		antennaBase(c, s, x, top)
		c.dc.DrawCircle(x, top-130, 34)
		c.fillOutlined(parseHex(s.Accent))
	},
	"double": func(c *canvas, s Spec, b Layout) {
		top := b.Y
		for _, side := range []float64{-1, 1} {
			x := b.CX() + side*70
			stem(c, x, top, x+side*50, top-110)
			antennaBase(c, s, x, top+8)
			c.dc.DrawCircle(x+side*58, top-128, 26)
			c.fillOutlined(parseHex(s.Accent))
		}
	},
	"bolt": func(c *canvas, s Spec, b Layout) {
		x, top := b.CX(), b.Y
		c.dc.MoveTo(x, top)
		c.dc.LineTo(x-35, top-70)
		c.dc.LineTo(x+30, top-90)
		c.dc.LineTo(x-10, top-170)
		c.outlinedStroke(parseHex(s.Accent), 18)
		antennaBase(c, s, x, top)
	},
}

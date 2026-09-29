package robot

func earY(b Layout) float64 { f := faceBox(b); return f.Y + f.H*0.5 }

var earParts = map[string]part{
	"none": func(*canvas, Spec, Layout) {},
	"bolts": func(c *canvas, s Spec, b Layout) {
		accent := parseHex(s.Accent)
		blocks, dots := boltEars(s, b)
		for _, r := range blocks {
			c.dc.DrawRoundedRectangle(r.x, r.y, r.w, r.h, r.r)
			c.fillOutlined(accent)
		}
		for _, d := range dots {
			c.dc.DrawCircle(d.x, d.y, d.r)
			c.dc.SetColor(ink)
			c.dc.Fill()
		}
	},
	"dials": func(c *canvas, s Spec, b Layout) {
		accent, body := parseHex(s.Accent), parseHex(s.Body)
		for _, d := range dialEars(s, b) {
			c.dc.DrawCircle(d.x, d.y, d.r)
			c.fillOutlined(accent)
			c.dc.DrawCircle(d.x, d.y, dialCapRadius)
			c.fillOutlined(body)
			c.dc.MoveTo(d.x, d.y)
			c.dc.LineTo(d.x, d.y-dialTick)
			c.strokeWith(ink, 8)
		}
	},
}

package robot

func earY(b Layout) float64 { f := faceBox(b); return f.Y + f.H*0.5 }

var earParts = map[string]part{
	"none": func(*canvas, Spec, Layout) {},
	"bolts": func(c *canvas, s Spec, b Layout) {
		accent, y := parseHex(s.Accent), earY(b)
		in := leftEdge(s.Head, b, y) - b.X
		for _, x := range []float64{b.X + in - 45, b.X + b.W - in - 25} {
			c.dc.DrawRoundedRectangle(x, y-50, 70, 100, 14)
			c.fillOutlined(accent)
		}
		for _, x := range []float64{b.X + in - 22, b.X + b.W - in + 22} {
			c.dc.DrawCircle(x, y, 12)
			c.dc.SetColor(ink)
			c.dc.Fill()
		}
	},
	"dials": func(c *canvas, s Spec, b Layout) {
		accent, body, y := parseHex(s.Accent), parseHex(s.Body), earY(b)
		in := leftEdge(s.Head, b, y) - b.X
		for _, x := range []float64{b.X + in - 25, b.X + b.W - in + 25} {
			c.dc.DrawCircle(x, y, 62)
			c.fillOutlined(accent)
			c.dc.DrawCircle(x, y, 30)
			c.fillOutlined(body)
			c.dc.MoveTo(x, y)
			c.dc.LineTo(x, y-30)
			c.strokeWith(ink, 8)
		}
	},
}

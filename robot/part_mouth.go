package robot

import "math"

func mouthY(b Layout) float64 { return b.Y + b.H*0.76 }

var mouthParts = map[string]part{
	"grille": func(c *canvas, s Spec, b Layout) {
		y := mouthY(b)
		c.dc.DrawRoundedRectangle(b.CX()-120, y-35, 240, 70, 20)
		c.fillOutlined(metal)
		for x := b.CX() - 80; x <= b.CX()+80; x += 40 {
			c.dc.MoveTo(x, y-35)
			c.dc.LineTo(x, y+35)
		}
		c.strokeWith(ink, 8)
	},
	"speaker": func(c *canvas, s Spec, b Layout) {
		x, y := b.CX(), mouthY(b)
		c.dc.DrawCircle(x, y, 52)
		c.fillOutlined(metal)
		c.dc.DrawCircle(x, y, 34)
		c.dc.DrawCircle(x, y, 18)
		c.strokeWith(ink, 6)
		c.dc.DrawCircle(x, y, 6)
		c.dc.SetColor(ink)
		c.dc.Fill()
	},
	"smile": func(c *canvas, s Spec, b Layout) {
		c.dc.DrawArc(b.CX(), mouthY(b)-60, 110, 0.2*math.Pi, 0.8*math.Pi)
		c.strokeWith(ink, 16)
	},
	"zigzag": func(c *canvas, s Spec, b Layout) {
		y := mouthY(b)
		for i := 0; i <= 6; i++ {
			x, dy := b.CX()-120+float64(i)*40, 20.0
			if i%2 == 1 {
				dy = -20
			}
			if i == 0 {
				c.dc.MoveTo(x, y+dy)
			} else {
				c.dc.LineTo(x, y+dy)
			}
		}
		c.strokeWith(ink, 14)
	},
}

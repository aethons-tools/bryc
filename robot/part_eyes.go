package robot

import "image/color"

const eyeGap = 110 // distance from the head's center line to each eye

// The widest and tallest eye styles, which bound how far the face can move.
const (
	visorHalfWidth = 190 // half the visor's width
	cyclopsRadius  = 95
)

func eyeY(s Spec, b Layout) float64 { ey, _ := facePos(s, b); return ey }

func eyeXs(b Layout) []float64 { return []float64{b.CX() - eyeGap, b.CX() + eyeGap} }

var eyeParts = map[string]part{
	"round": func(c *canvas, s Spec, b Layout) {
		glow, y := parseHex(s.Glow), eyeY(s, b)
		for _, x := range eyeXs(b) {
			halo(c, x, y, 58, glow)
			c.dc.DrawCircle(x, y, 58)
			c.fillOutlined(glow)
			highlight(c, x-18, y-18, 16)
		}
	},
	"visor": func(c *canvas, s Spec, b Layout) {
		glow, y := parseHex(s.Glow), eyeY(s, b)
		c.dc.DrawRoundedRectangle(b.CX()-visorHalfWidth, y-55, 2*visorHalfWidth, 110, 55)
		c.fillOutlined(screen)
		c.dc.DrawRoundedRectangle(b.CX()-160, y-22, 320, 44, 22)
		c.dc.SetColor(glow)
		c.dc.Fill()
		highlight(c, b.CX()-130, y-10, 9)
	},
	"cyclops": func(c *canvas, s Spec, b Layout) {
		glow, x, y := parseHex(s.Glow), b.CX(), eyeY(s, b)
		halo(c, x, y, cyclopsRadius, glow)
		c.dc.DrawCircle(x, y, cyclopsRadius)
		c.fillOutlined(glow)
		c.dc.DrawCircle(x, y, 40)
		c.dc.SetColor(shade(glow, 0.55))
		c.dc.Fill()
		highlight(c, x-30, y-30, 20)
	},
	"led": func(c *canvas, s Spec, b Layout) {
		glow, y := parseHex(s.Glow), eyeY(s, b)
		for _, x := range eyeXs(b) {
			halo(c, x, y, 70, glow)
			c.dc.DrawRectangle(x-55, y-55, 110, 110)
			c.fillOutlined(glow)
			c.dc.MoveTo(x, y-55)
			c.dc.LineTo(x, y+55)
			c.dc.MoveTo(x-55, y)
			c.dc.LineTo(x+55, y)
			c.strokeWith(ink, 6)
		}
	},
}

// halo paints a soft glow behind an eye as stacked translucent discs.
func halo(c *canvas, x, y, r float64, glow color.NRGBA) {
	for i := 3; i >= 1; i-- {
		c.dc.DrawCircle(x, y, r+float64(i)*14)
		c.dc.SetColor(withAlpha(glow, 40))
		c.dc.Fill()
	}
}

// highlight adds the white shine dot that makes eyes look glassy.
func highlight(c *canvas, x, y, r float64) {
	c.dc.DrawCircle(x, y, r)
	c.dc.SetColor(color.NRGBA{255, 255, 255, 200})
	c.dc.Fill()
}

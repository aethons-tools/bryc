package robot

import "image/color"

const eyeGap = 110 // distance from the head's center line to each eye

const roundEyeRadius = 58

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
			c.dc.DrawCircle(x, y, roundEyeRadius)
			c.fillOutlined(glow)
			// The glow goes over the outline, like a lit bulb.
			halo(c, x, y, roundEyeRadius, glow)
			hotSpot(c, x, y, glow)
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
	"cyclops": drawHAL,
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

// Cyclops (HAL-style) geometry: a metal ring around a black lens with the
// glow lit in its center.
const (
	cyclopsRing     = 12.0 // width of the metal ring inside the outline
	cyclopsCore     = 18.0 // radius of the solid glow core
	cyclopsGlowFade = 60.0 // radius the glow fades out by
)

func drawHAL(c *canvas, s Spec, b Layout) {
	glow, x, y := parseHex(s.Glow), b.CX(), eyeY(s, b)
	c.dc.DrawCircle(x, y, cyclopsRadius)
	c.fillOutlined(metal)
	c.dc.DrawCircle(x, y, cyclopsRadius-cyclopsRing)
	c.dc.SetColor(screen)
	c.dc.FillPreserve()
	c.dc.SetColor(ink)
	c.lw(5)
	c.dc.Stroke()
	// The glow fades out from the core in translucent steps.
	for r := cyclopsGlowFade; r > cyclopsCore; r -= 10 {
		c.dc.DrawCircle(x, y, r)
		c.dc.SetColor(withAlpha(glow, 45))
		c.dc.Fill()
	}
	c.dc.DrawCircle(x, y, cyclopsCore)
	c.dc.SetColor(glow)
	c.dc.Fill()
	hotSpot(c, x, y, glow)
	// A faint glint on the glass.
	c.dc.DrawCircle(x-34, y-34, 14)
	c.dc.SetColor(color.NRGBA{255, 255, 255, 110})
	c.dc.Fill()
}

// hotSpotRadius is the pale center of a lit eye (cyclops and round eyes).
const hotSpotRadius = 7.0

func hotSpotColor(glow color.NRGBA) color.NRGBA {
	return mix(glow, color.NRGBA{255, 255, 255, 255}, 0.6)
}

// hotSpot paints the pale center of a lit eye.
func hotSpot(c *canvas, x, y float64, glow color.NRGBA) {
	c.dc.DrawCircle(x, y, hotSpotRadius)
	c.dc.SetColor(hotSpotColor(glow))
	c.dc.Fill()
}

// halo paints a soft glow around an eye as stacked translucent discs.
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

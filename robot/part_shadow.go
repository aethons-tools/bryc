package robot

import "image/color"

// shadowLayers is how many nested ellipses build the shadow's soft edge.
const shadowLayers = 10

// drawShadow draws a soft oval shadow on the floor below the floating head:
// nested ellipses of faint ink, so the middle is darkest and the edge fades.
func drawShadow(c *canvas, head Layout) {
	cx, cy, rx, ry := floorShadow(head)
	c.dc.SetColor(color.NRGBA{ink.R, ink.G, ink.B, 0x0a})
	for i := range shadowLayers {
		k := 1 - 0.5*float64(i)/shadowLayers
		c.dc.DrawEllipse(cx, cy, rx*k, ry*k)
		c.dc.Fill()
	}
}

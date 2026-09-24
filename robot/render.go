package robot

import (
	"image"

	"github.com/fogleman/gg"
)

// Render draws a resolved, validated Spec as a size×size image.
func Render(s Spec, size int) image.Image {
	c := &canvas{dc: gg.NewContext(size, size), k: float64(size) / virtual}
	c.dc.Scale(c.k, c.k)
	if s.Background != "none" {
		c.dc.SetColor(parseHex(s.Background))
		c.dc.Clear()
	}
	drawBody(c, s, headBox)
	drawHead(c, s, headBox)
	return c.dc.Image()
}

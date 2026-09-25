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
	head := headLayout(s)
	drawBody(c, s, head)
	earParts[s.Ears](c, s, head)
	drawHead(c, s, head)
	if *s.Panels {
		drawPanels(c, s, head)
	}
	if *s.Rivets {
		drawRivets(c, s, head)
	}
	eyeParts[s.Eyes](c, s, head)
	mouthParts[s.Mouth](c, s, head)
	if *s.Blush {
		drawBlush(c, s, head)
	}
	antennaParts[s.Antenna](c, s, head)
	return c.dc.Image()
}

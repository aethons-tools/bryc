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
	earParts[s.Ears](c, s, headBox)
	drawHead(c, s, headBox)
	if *s.Panels {
		drawPanels(c, s, headBox)
	}
	if *s.Rivets {
		drawRivets(c, s, headBox)
	}
	eyeParts[s.Eyes](c, s, headBox)
	mouthParts[s.Mouth](c, s, headBox)
	if *s.Blush {
		drawBlush(c, s, headBox)
	}
	antennaParts[s.Antenna](c, s, headBox)
	return c.dc.Image()
}

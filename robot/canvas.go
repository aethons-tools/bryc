package robot

import (
	"image/color"

	"github.com/fogleman/gg"
)

// All drawing uses a virtual 1000×1000 space; Render scales it to the output.
const virtual = 1000.0

// outline is the width of the dark cartoon outline around every shape.
const outline = 14.0

var (
	ink    = color.NRGBA{0x26, 0x26, 0x2e, 0xff} // outlines and dark details
	metal  = color.NRGBA{0xd9, 0xdd, 0xe3, 0xff} // grilles, bolts, rivets
	screen = color.NRGBA{0x1d, 0x1d, 0x24, 0xff} // visor background
)

// Layout is a box in virtual units.
type Layout struct{ X, Y, W, H float64 }

// CX is the horizontal center of the box.
func (b Layout) CX() float64 { return b.X + b.W/2 }

// headBox is where every head shape is drawn; other parts are placed
// relative to it so all combinations line up.
var headBox = Layout{X: 250, Y: 260, W: 500, H: 460}

// part draws one facet variant.
type part func(c *canvas, s Spec, b Layout)

// canvas is a gg context whose user space is the virtual square. gg does not
// scale line widths with its transform, so widths must go through lw.
type canvas struct {
	dc *gg.Context
	k  float64 // output pixels per virtual unit
}

func (c *canvas) lw(w float64) { c.dc.SetLineWidth(w * c.k) }

// fillOutlined fills the current path, then strokes it with the ink outline.
func (c *canvas) fillOutlined(fill color.Color) {
	c.dc.SetColor(fill)
	c.dc.FillPreserve()
	c.dc.SetColor(ink)
	c.lw(outline)
	c.dc.SetLineJoinRound()
	c.dc.Stroke()
}

// strokeWith strokes the current path with round caps and joins.
func (c *canvas) strokeWith(col color.Color, w float64) {
	c.dc.SetColor(col)
	c.lw(w)
	c.dc.SetLineCapRound()
	c.dc.SetLineJoinRound()
	c.dc.Stroke()
}

// outlinedStroke strokes the current path as a colored line with an ink border.
func (c *canvas) outlinedStroke(col color.Color, w float64) {
	c.dc.SetLineCapRound()
	c.dc.SetLineJoinRound()
	c.dc.SetColor(ink)
	c.lw(w + outline)
	c.dc.StrokePreserve()
	c.strokeWith(col, w)
}

package robot

// drawPanels draws two seam lines across the head, clipped to its outline.
func drawPanels(c *canvas, s Spec, b Layout) {
	headPaths[s.Head](c.dc, b)
	c.dc.Clip()
	c.dc.MoveTo(b.X, b.Y+90)
	c.dc.LineTo(b.X+b.W, b.Y+90)
	c.dc.MoveTo(b.X, b.Y+b.H-30)
	c.dc.LineTo(b.X+b.W, b.Y+b.H-30)
	c.strokeWith(shade(parseHex(s.Body), 0.6), 8)
	c.dc.ResetClip()
}

// drawRivets places four rivets on the lower half of the head, where every
// head shape is wide enough to hold them.
func drawRivets(c *canvas, s Spec, b Layout) {
	for _, y := range []float64{b.Y + b.H*0.55, b.Y + b.H - 50} {
		for _, x := range []float64{b.X + 45, b.X + b.W - 45} {
			c.dc.DrawCircle(x, y, 12)
			c.dc.SetColor(metal)
			c.dc.FillPreserve()
			c.dc.SetColor(ink)
			c.lw(5)
			c.dc.Stroke()
		}
	}
}

// drawBlush adds translucent accent-colored cheeks.
func drawBlush(c *canvas, s Spec, b Layout) {
	for _, x := range []float64{b.CX() - 170, b.CX() + 170} {
		c.dc.DrawEllipse(x, blushY(s, b), 34, 20)
		c.dc.SetColor(withAlpha(parseHex(s.Accent), 150))
		c.dc.Fill()
	}
}

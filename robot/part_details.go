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

const (
	rivetRadius = 12.0
	rivetInset  = 45.0 // distance in from the head's edge
	// maxRivetEdge is how far in from the normal head box's side the head's
	// edge may be at the lower rivets' height; on heads that narrow more than
	// this toward the bottom, the lower rivets move up instead of crowding
	// the mouth, which is always at the center.
	maxRivetEdge = 60.0
)

// rivetCenters places four rivets on the lower part of the head, set in from
// its edge so they stay on heads that narrow toward the bottom.
func rivetCenters(s Spec, b Layout) [][2]float64 {
	f := faceBox(b)
	lower := f.Y + f.H - 50
	for lower > f.Y+f.H*0.55 && leftEdge(s.Head, b, lower) > headBox.X+maxRivetEdge {
		lower -= 2
	}
	var centers [][2]float64
	for _, y := range []float64{f.Y + f.H*0.55, lower} {
		in := leftEdge(s.Head, b, y) - b.X + rivetInset
		centers = append(centers, [2]float64{b.X + in, y}, [2]float64{b.X + b.W - in, y})
	}
	return centers
}

func drawRivets(c *canvas, s Spec, b Layout) {
	for _, p := range rivetCenters(s, b) {
		c.dc.DrawCircle(p[0], p[1], rivetRadius)
		c.dc.SetColor(metal)
		c.dc.FillPreserve()
		c.dc.SetColor(ink)
		c.lw(5)
		c.dc.Stroke()
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

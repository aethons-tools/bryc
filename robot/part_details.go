package robot

// panelShade is the seam color relative to the body color.
const panelShade = 0.6

// panelSeam is one seam: its height where it meets the head's sides, and
// which way it bows on heads that bow (-1 up, like the top edge; +1 down).
type panelSeam struct{ y, dir float64 }

func panelSeams(s Spec, b Layout) []panelSeam {
	return []panelSeam{{b.Y + 90, -1}, {b.Y + b.H - 30, +1}}
}

// drawPanels draws two seam lines across the head, clipped to its outline.
// On heads whose edges bow (see bowTo), the seams bow the same way as the
// nearest edge, by the same fraction of the head's width at their height.
func drawPanels(c *canvas, s Spec, b Layout) {
	headPaths[s.Head](c.dc, b)
	c.dc.Clip()
	for _, seam := range panelSeams(s, b) {
		// Run from edge to edge (a little past, the clip trims it) so the bow
		// is sized to the head's width at this height.
		x0 := leftEdge(s.Head, b, seam.y)
		x1 := 2*b.CX() - x0
		c.dc.MoveTo(x0, seam.y)
		if headBows[s.Head] {
			bowTo(c.dc, x0, x1, seam.y, seam.dir)
		} else {
			c.dc.LineTo(x1, seam.y)
		}
	}
	c.strokeWith(shade(parseHex(s.Body), panelShade), 8)
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

package robot

import "math"

// Shoulder geometry, in virtual units. The shoulders run off the bottom edge.
const (
	shoulderLeft, shoulderRight = 130.0, 870.0
	shoulderTop, shoulderBottom = 830.0, 1430.0
	maxShoulderRadius           = 280.0 // corner radius at shoulders=100
	maxSpikeHeight              = 150.0 // spike rise at shoulders=-100
	maxSpikeLean                = 40.0  // how far the tip leans outward at -100
	spikeBase                   = 160.0 // width of each spike along the shoulder top
)

// drawBody draws the neck, the shoulders, and an accent-colored chest light.
func drawBody(c *canvas, s Spec, head Layout) {
	body := parseHex(s.Body)
	c.dc.DrawRectangle(head.CX()-70, head.Y+head.H-30, 140, 150)
	c.fillOutlined(shade(body, 0.7))
	shoulderPath(c, *s.Shoulders)
	c.fillOutlined(body)
	c.dc.DrawCircle(head.CX(), 920, 30)
	c.fillOutlined(parseHex(s.Accent))
}

// shoulderPath traces the shoulders as one closed path so the outline is
// continuous: 0 is a plain rectangle, positive values round the top corners
// (radius grows with the value), negative values raise a spike on each
// shoulder (height grows with the magnitude).
func shoulderPath(c *canvas, shoulders int) {
	dc := c.dc
	t := float64(shoulders) / ShouldersMax
	dc.NewSubPath()
	dc.MoveTo(shoulderLeft, shoulderBottom)
	if t >= 0 {
		r := maxShoulderRadius * t
		dc.LineTo(shoulderLeft, shoulderTop+r)
		if r > 0 {
			dc.DrawArc(shoulderLeft+r, shoulderTop+r, r, math.Pi, 1.5*math.Pi)
		}
		dc.LineTo(shoulderRight-r, shoulderTop)
		if r > 0 {
			dc.DrawArc(shoulderRight-r, shoulderTop+r, r, 1.5*math.Pi, 2*math.Pi)
		}
	} else {
		h, lean := maxSpikeHeight*-t, maxSpikeLean*-t
		dc.LineTo(shoulderLeft, shoulderTop)
		dc.LineTo(shoulderLeft-lean, shoulderTop-h)
		dc.LineTo(shoulderLeft+spikeBase, shoulderTop)
		dc.LineTo(shoulderRight-spikeBase, shoulderTop)
		dc.LineTo(shoulderRight+lean, shoulderTop-h)
		dc.LineTo(shoulderRight, shoulderTop)
	}
	dc.LineTo(shoulderRight, shoulderBottom)
	dc.ClosePath()
}

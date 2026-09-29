package robot

// drawAntenna draws s's antenna (see antennaGeometry): the zigzag bolt, then
// each stem, its mount and the ball on its tip.
func drawAntenna(c *canvas, s Spec, b Layout) {
	a := antennaGeometry(s, b)
	if len(a.bolt) > 0 {
		c.dc.MoveTo(a.bolt[0][0], a.bolt[0][1])
		for _, p := range a.bolt[1:] {
			c.dc.LineTo(p[0], p[1])
		}
		c.outlinedStroke(parseHex(s.Accent), 18)
	}
	for i, base := range a.bases {
		if i < len(a.stems) {
			st := a.stems[i]
			c.dc.MoveTo(st[0][0], st[0][1])
			c.dc.LineTo(st[1][0], st[1][1])
			c.strokeWith(ink, antennaStemWidth)
		}
		c.dc.DrawRoundedRectangle(base.x, base.y, base.w, base.h, base.r)
		c.fillOutlined(shade(parseHex(s.Body), 0.8))
		if i < len(a.balls) {
			ball := a.balls[i]
			c.dc.DrawCircle(ball.x, ball.y, ball.r)
			c.fillOutlined(parseHex(s.Accent))
		}
	}
}

var antennaParts = map[string]part{
	"none":   func(*canvas, Spec, Layout) {},
	"ball":   drawAntenna,
	"double": drawAntenna,
	"bolt":   drawAntenna,
}

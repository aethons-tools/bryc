package robot

// drawBody draws the neck, the shoulders running off the bottom edge, and an
// accent-colored chest light.
func drawBody(c *canvas, s Spec, head Layout) {
	body := parseHex(s.Body)
	c.dc.DrawRectangle(head.CX()-70, head.Y+head.H-30, 140, 150)
	c.fillOutlined(shade(body, 0.7))
	c.dc.DrawRoundedRectangle(130, 830, 740, 300, 120)
	c.fillOutlined(body)
	c.dc.DrawCircle(head.CX(), 920, 30)
	c.fillOutlined(parseHex(s.Accent))
}

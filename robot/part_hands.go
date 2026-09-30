package robot

import "github.com/aethons-tools/bryc/robot/mesh"

// drawHands draws the two floating hands in the body color. Each digit is
// drawn before the palm, so the palm's outline covers their roots.
func drawHands(c *canvas, s Spec, b Layout) {
	body := parseHex(s.Body)
	for _, h := range hands(b, handTilt2D) {
		for _, d := range h.digits {
			c.polygon(h, capsuleOutline(d.root, d.tip, d.r))
			c.fillOutlined(body)
		}
		c.polygon(h, h.palm)
		c.fillOutlined(body)
	}
}

// polygon traces the closed polygon of wrist offsets pts for hand h.
func (c *canvas) polygon(h hand, pts []mesh.Vec2) {
	c.dc.NewSubPath()
	for _, p := range pts {
		q := h.at(p)
		c.dc.LineTo(q.X, q.Y)
	}
	c.dc.ClosePath()
}

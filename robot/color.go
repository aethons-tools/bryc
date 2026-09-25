package robot

import (
	"fmt"
	"image/color"
	"math"
	"strconv"
)

// parseHex parses "#rrggbb". Callers only pass validated values, so a
// malformed color is a programming error.
func parseHex(s string) color.NRGBA {
	n, err := strconv.ParseUint(s[1:], 16, 32)
	if len(s) != 7 || err != nil {
		panic(fmt.Sprintf("robot: invalid color %q", s))
	}
	return color.NRGBA{uint8(n >> 16), uint8(n >> 8), uint8(n), 0xff}
}

func hexString(c color.NRGBA) string {
	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
}

// hsl converts hue in degrees [0,360) and saturation/lightness in [0,1].
func hsl(h, s, l float64) color.NRGBA {
	c := (1 - math.Abs(2*l-1)) * s
	hp := h / 60
	x := c * (1 - math.Abs(math.Mod(hp, 2)-1))
	var r, g, b float64
	switch {
	case hp < 1:
		r, g, b = c, x, 0
	case hp < 2:
		r, g, b = x, c, 0
	case hp < 3:
		r, g, b = 0, c, x
	case hp < 4:
		r, g, b = 0, x, c
	case hp < 5:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}
	m := l - c/2
	to := func(v float64) uint8 { return uint8(math.Round((v + m) * 255)) }
	return color.NRGBA{to(r), to(g), to(b), 0xff}
}

// shade multiplies RGB by f (f<1 darkens, f>1 lightens), clamping at 255.
func shade(c color.NRGBA, f float64) color.NRGBA {
	ch := func(v uint8) uint8 { return uint8(math.Min(255, math.Round(float64(v)*f))) }
	return color.NRGBA{ch(c.R), ch(c.G), ch(c.B), c.A}
}

func withAlpha(c color.NRGBA, a uint8) color.NRGBA {
	c.A = a
	return c
}

// mix blends a toward b by t in [0,1].
func mix(a, b color.NRGBA, t float64) color.NRGBA {
	ch := func(x, y uint8) uint8 { return uint8(math.Round(float64(x) + (float64(y)-float64(x))*t)) }
	return color.NRGBA{ch(a.R, b.R), ch(a.G, b.G), ch(a.B, b.B), ch(a.A, b.A)}
}

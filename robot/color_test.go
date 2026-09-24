package robot

import (
	"image/color"
	"testing"
)

func TestHSL(t *testing.T) {
	cases := []struct {
		h, s, l float64
		want    string
	}{
		{0, 1, 0.5, "#ff0000"},
		{120, 1, 0.5, "#00ff00"},
		{240, 1, 0.25, "#000080"},
		{0, 0, 1, "#ffffff"},
		{300, 1, 0.5, "#ff00ff"},
	}
	for _, c := range cases {
		if got := hexString(hsl(c.h, c.s, c.l)); got != c.want {
			t.Errorf("hsl(%v,%v,%v) = %s, want %s", c.h, c.s, c.l, got, c.want)
		}
	}
}

func TestParseHexRoundTrip(t *testing.T) {
	c := parseHex("#1a2b3c")
	if c != (color.NRGBA{0x1a, 0x2b, 0x3c, 0xff}) {
		t.Errorf("parseHex = %v", c)
	}
	if hexString(c) != "#1a2b3c" {
		t.Errorf("hexString = %s", hexString(c))
	}
}

func TestShadeClamps(t *testing.T) {
	c := color.NRGBA{200, 100, 0, 255}
	if got := shade(c, 0.5); got != (color.NRGBA{100, 50, 0, 255}) {
		t.Errorf("darken = %v", got)
	}
	if got := shade(c, 2); got != (color.NRGBA{255, 200, 0, 255}) {
		t.Errorf("lighten = %v", got)
	}
}

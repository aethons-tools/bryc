package robot

import "image/color"

const eyeGap = 110 // distance from the head's center line to each eye

// roundEyeRadius is the size of a round eye when there is a single pair.
const roundEyeRadius = 58

// roundLayout arranges round eyes: a single centered eye, or stacked pairs.
// It holds the largest radius the layout allows, the vertical distance
// between rows, and each column's distance from the head's center line
// (0 for a single eye). More eyes are smaller so the group is never taller
// than the largest single eye or wider than the visor.
type roundLayout struct {
	r, rowGap, colGap float64
	rows              int
}

var roundLayouts = map[string]roundLayout{
	"1": {r: roundEyeSizes["4"], rows: 1},
	"2": {r: roundEyeSizes["3"], colGap: eyeGap, rows: 1},
	"4": {r: roundEyeSizes["2"], rowGap: 104, colGap: 100, rows: 2},
	"6": {r: roundEyeSizes["1"], rowGap: 68, colGap: 88, rows: 3},
}

// roundEyeSizes is the radius for each eye size: the sizes the round-eye
// layouts use for 6, 4, 2 and 1 eyes.
var roundEyeSizes = map[string]float64{"1": 25, "2": 40, "3": roundEyeRadius, "4": maxEyeRadius}

// roundEyeR is the round-eye radius: the eye size, capped at the largest the
// eye count's layout allows. Smaller eyes keep the layout's positions.
func roundEyeR(s Spec) float64 {
	return min(roundEyeSizes[s.EyeSize], roundLayouts[s.EyeCount].r)
}

// roundEyeCenters returns the center of every round eye, rows centered on
// the face's eye line.
func roundEyeCenters(s Spec, b Layout) [][2]float64 {
	l := roundLayouts[s.EyeCount]
	top := eyeY(s, b) - float64(l.rows-1)*l.rowGap/2
	var centers [][2]float64
	for row := range l.rows {
		y := top + float64(row)*l.rowGap
		if l.colGap == 0 {
			centers = append(centers, [2]float64{b.CX(), y})
			continue
		}
		centers = append(centers, [2]float64{b.CX() - l.colGap, y}, [2]float64{b.CX() + l.colGap, y})
	}
	return centers
}

// The widest and tallest eyes, which bound how far the face can move: the
// visor, and a single round eye at its largest.
const (
	visorHalfWidth = 180 // half the visor's width
	maxEyeRadius   = 95
)

func eyeY(s Spec, b Layout) float64 { ey, _ := facePos(s, b); return ey }

var eyeParts = map[string]part{
	"round": func(c *canvas, s Spec, b Layout) {
		glow, r := parseHex(s.Glow), roundEyeR(s)
		draw := eyeStyles[s.EyeStyle]
		for _, p := range roundEyeCenters(s, b) {
			draw(c, p[0], p[1], r, glow)
		}
	},
	"visor": func(c *canvas, s Spec, b Layout) {
		glow, y := parseHex(s.Glow), eyeY(s, b)
		c.dc.DrawRoundedRectangle(b.CX()-visorHalfWidth, y-55, 2*visorHalfWidth, 110, 55)
		c.fillOutlined(screen)
		bar := visorHalfWidth - 30.0 // the glow bar sits 30 in from the visor's ends
		c.dc.DrawRoundedRectangle(b.CX()-bar, y-22, 2*bar, 44, 22)
		c.dc.SetColor(glow)
		c.dc.Fill()
		highlight(c, b.CX()-bar+30, y-10, 9)
	},
}

// eyeStyles draw one round eye of radius r centered at (x, y). Every detail
// scales with the eye, so small eyes look like smaller versions of big ones.
var eyeStyles = map[string]func(c *canvas, x, y, r float64, glow color.NRGBA){
	"bright": drawBrightEye,
	"glower": drawGlowerEye,
	"dead":   drawDeadEye,
}

// drawBrightEye is a lit bulb: glow fill, with the glow spilling over the
// outline, a pale hot spot and a glint.
func drawBrightEye(c *canvas, x, y, r float64, glow color.NRGBA) {
	k := r / roundEyeRadius
	c.dc.DrawCircle(x, y, r)
	c.fillOutlined(glow)
	halo(c, x, y, r, haloSpread*k, glow)
	hotSpot(c, x, y, hotSpotRadius*k, glow)
	highlight(c, x-18*k, y-18*k, 16*k)
}

// Glowering (HAL-style) eye geometry at the largest eye size: a metal ring
// around a black lens with the glow lit in its center.
const (
	glowerRing = 12.0 // width of the metal ring inside the outline
	glowerCore = 18.0 // radius of the solid glow core
	glowerFade = 60.0 // radius the glow fades out by
)

func drawGlowerEye(c *canvas, x, y, r float64, glow color.NRGBA) {
	k := r / maxEyeRadius
	c.dc.DrawCircle(x, y, r)
	c.fillOutlined(metal)
	c.dc.DrawCircle(x, y, r-glowerRing*k)
	c.dc.SetColor(screen)
	c.dc.FillPreserve()
	c.dc.SetColor(ink)
	c.lw(5 * k)
	c.dc.Stroke()
	// The glow fades out from the core in translucent steps.
	for fr := glowerFade * k; fr > glowerCore*k; fr -= 10 * k {
		c.dc.DrawCircle(x, y, fr)
		c.dc.SetColor(withAlpha(glow, 45))
		c.dc.Fill()
	}
	c.dc.DrawCircle(x, y, glowerCore*k)
	c.dc.SetColor(glow)
	c.dc.Fill()
	hotSpot(c, x, y, hotSpotRadius*k, glow)
	// A faint glint on the glass.
	c.dc.DrawCircle(x-34*k, y-34*k, 14*k)
	c.dc.SetColor(color.NRGBA{255, 255, 255, 110})
	c.dc.Fill()
}

// drawDeadEye is an unlit lens: black, with only the glint.
func drawDeadEye(c *canvas, x, y, r float64, _ color.NRGBA) {
	k := r / roundEyeRadius
	c.dc.DrawCircle(x, y, r)
	c.fillOutlined(screen)
	highlight(c, x-18*k, y-18*k, 16*k)
}

// hotSpotRadius is the pale center of a lit eye, for a full-size eye.
const hotSpotRadius = 7.0

func hotSpotColor(glow color.NRGBA) color.NRGBA {
	return mix(glow, color.NRGBA{255, 255, 255, 255}, 0.6)
}

// hotSpot paints the pale center of a lit eye.
func hotSpot(c *canvas, x, y, r float64, glow color.NRGBA) {
	c.dc.DrawCircle(x, y, r)
	c.dc.SetColor(hotSpotColor(glow))
	c.dc.Fill()
}

// haloSpread is how far each of a halo's three glow rings reaches past the
// last, for a full-size eye.
const haloSpread = 14.0

// halo paints a soft glow around an eye of radius r as stacked translucent
// discs, each reaching spread further out.
func halo(c *canvas, x, y, r, spread float64, glow color.NRGBA) {
	for i := 3; i >= 1; i-- {
		c.dc.DrawCircle(x, y, r+float64(i)*spread)
		c.dc.SetColor(withAlpha(glow, 40))
		c.dc.Fill()
	}
}

// highlight adds the white shine dot that makes eyes look glassy.
func highlight(c *canvas, x, y, r float64) {
	c.dc.DrawCircle(x, y, r)
	c.dc.SetColor(color.NRGBA{255, 255, 255, 200})
	c.dc.Fill()
}

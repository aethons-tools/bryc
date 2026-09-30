package robot

// Shared part geometry: where each part sits and how big it is, in canvas
// coordinates (y down). The 2D renderer draws from these, and the 3D model
// builds its meshes from the same numbers, so both agree.

// rect is a rounded rectangle by its top-left corner, size and corner radius.
type rect struct{ x, y, w, h, r float64 }

// circle is a circle by its center and radius.
type circle struct{ x, y, r float64 }

// floorShadow is the soft shadow on the floor under the floating head, as
// an ellipse: center, and horizontal and vertical radii.
func floorShadow(head Layout) (cx, cy, rx, ry float64) {
	return head.CX(), floorShadowY, head.W * 0.42, 26
}

// floorShadowY is the height of the floor the head floats over. The 3D
// viewer puts its shadow at the same height (1000−950 = 0.05 m).
const floorShadowY = 950.0

// boltEars are the "bolts" ears: a block on each side of the head with a
// dark dot where it meets the head, left first.
func boltEars(s Spec, b Layout) (blocks [2]rect, dots [2]circle) {
	y := earY(b)
	in := leftEdge(s.Head, b, y) - b.X
	for i, x := range []float64{b.X + in - 45, b.X + b.W - in - 25} {
		blocks[i] = rect{x, y - 50, 70, 100, 14}
	}
	for i, x := range []float64{b.X + in - 22, b.X + b.W - in + 22} {
		dots[i] = circle{x, y, 12}
	}
	return blocks, dots
}

// Dial ears: the dial, its cap, and the length of the tick on the cap.
const dialRadius, dialCapRadius, dialTick = 62.0, 30.0, 30.0

// dialEars are the "dials" ears' dials, left first.
func dialEars(s Spec, b Layout) [2]circle {
	y := earY(b)
	in := leftEdge(s.Head, b, y) - b.X
	var dials [2]circle
	for i, x := range []float64{b.X + in - 25, b.X + b.W - in + 25} {
		dials[i] = circle{x, y, dialRadius}
	}
	return dials
}

// antennaStemWidth is the width of an antenna's stem.
const antennaStemWidth = 14.0

// antenna is an antenna's parts: straight stems (from, to), balls on their
// tips, the mounts where they meet the head, and the zigzag bolt's points.
// Parts with the same index belong together and are drawn stem, base, ball.
type antenna struct {
	stems [][2][2]float64
	balls []circle
	bases []rect
	bolt  [][2]float64
}

// antennaBaseAt is the mount for an antenna meeting the head at (x, y).
func antennaBaseAt(x, y float64) rect { return rect{x - 45, y - 22, 90, 34, 10} }

// antennaGeometry is s's antenna on the head in b.
func antennaGeometry(s Spec, b Layout) antenna {
	var a antenna
	switch s.Antenna {
	case "ball":
		x, top := b.CX(), b.Y
		a.stems = append(a.stems, [2][2]float64{{x, top}, {x, top - 110}})
		a.bases = append(a.bases, antennaBaseAt(x, top))
		a.balls = append(a.balls, circle{x, top - 130, 34})
	case "double":
		top := b.Y
		for _, side := range []float64{-1, 1} {
			x := b.CX() + side*70
			a.stems = append(a.stems, [2][2]float64{{x, top}, {x + side*50, top - 110}})
			a.bases = append(a.bases, antennaBaseAt(x, top+8))
			a.balls = append(a.balls, circle{x + side*58, top - 128, 26})
		}
	case "bolt":
		x, top := b.CX(), b.Y
		a.bolt = [][2]float64{{x, top}, {x - 35, top - 70}, {x + 30, top - 90}, {x - 10, top - 170}}
		a.bases = append(a.bases, antennaBaseAt(x, top))
	}
	return a
}

// Visor geometry: the visor's half height (its ends are half circles), and
// the glow bar inside it, set in from the visor's ends.
const (
	visorHalfHeight    = 55.0
	visorBarInset      = 30.0
	visorBarHalfHeight = 22.0
)

// Mouth sizes: the grille's width, the slot's size, and the line's width.
const (
	grilleWidth           = 240.0
	slotWidth, slotHeight = 120.0, 36.0
	lineMouthWidth        = 180.0
)

// grilleBars is the x of each bar across a grille centered at cx.
func grilleBars(cx float64) []float64 {
	var xs []float64
	for x := cx - 80; x <= cx+80; x += 40 {
		xs = append(xs, x)
	}
	return xs
}

// Blush cheeks are ellipses of these radii.
const blushRX, blushRY = 34.0, 20.0

// blushCenters are the centers of the blush cheeks, left first.
func blushCenters(s Spec, b Layout) [2][2]float64 {
	y := blushY(s, b)
	return [2][2]float64{{b.CX() - 170, y}, {b.CX() + 170, y}}
}

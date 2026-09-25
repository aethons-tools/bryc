package robot

// Face layout, as fractions of the face box (see faceBox) from its top. Only the
// eyes, mouth (including the jaw) and blush move with the face position;
// head hardware (ears, antenna, rivets, panels) stays put.
//
//	-1  eyes and mouth move down, closer together
//	 0  neutral
//	+1  eyes rise toward the top of the head; the mouth stays put
//	+2  eyes as at +1; the mouth rises to sit close under them
const (
	neutralEyes   = 0.40
	neutralMouth  = 0.76
	loweredEyes   = 0.50 // face=-1
	loweredMouth  = 0.82 // face=-1
	tightMouthGap = 0.35 // face=+2: mouth this far below the eyes, clear of the cyclops
	blushBetween  = 0.70 // blush sits this far from the eyes toward the mouth
)

// raisedEyes is the eye height at face=+1 and +2 for each head shape. Heads
// that narrow toward the top (dome, trapezoid) rise less, so the widest eye
// style (the visor) and the tallest (the cyclops) always fit inside.
var raisedEyes = map[string]float64{
	"square":             0.3025,
	"rounded":            0.3025,
	"dome":               0.3475,
	"trapezoid":          0.3475,
	"inverted-dome":      0.3025, // flat, full-width top
	"inverted-trapezoid": 0.3025, // widest at the top
}

// facePos returns the eye and mouth center heights for s's face position.
func facePos(s Spec, b Layout) (eyeY, mouthY float64) {
	eyes, mouth := neutralEyes, neutralMouth
	switch *s.Face {
	case -1:
		eyes, mouth = loweredEyes, loweredMouth
	case 1:
		eyes = raisedEyes[s.Head]
	case 2:
		eyes = raisedEyes[s.Head]
		mouth = eyes + tightMouthGap
	}
	f := faceBox(b)
	return f.Y + f.H*eyes, f.Y + f.H*mouth
}

// blushY is the height of the cheek blush, between the eyes and the mouth.
func blushY(s Spec, b Layout) float64 {
	ey, my := facePos(s, b)
	return ey + (my-ey)*blushBetween
}

package robot

// Face layout, as fractions of the head box height from its top. Only the
// eyes, mouth (including the jaw) and blush move with the face position;
// head hardware (ears, antenna, rivets, panels) stays put.
const (
	neutralEyes   = 0.40
	neutralMouth  = 0.76
	squashedEyes  = 0.60 // face=-2
	squashedMouth = 0.88 // face=-2
	blushBetween  = 0.70 // blush sits this far from the eyes toward the mouth
)

// raisedEyes is the eye height at face=+2 for each head shape: as high as
// both the widest eye style (the visor) and the tallest (the cyclops) fit
// inside the head, so heads that narrow toward the top (dome, trapezoid)
// stretch less than square ones.
var raisedEyes = map[string]float64{
	"square":    0.27,
	"rounded":   0.27,
	"dome":      0.33,
	"trapezoid": 0.33,
}

// facePos returns the eye and mouth center heights for s's face position.
// Positive values raise the eyes toward the top of the head and leave the
// mouth where it is; negative values move both down, closing the gap.
func facePos(s Spec, b Layout) (eyeY, mouthY float64) {
	t := float64(*s.Face) / FaceMax
	eyes, mouth := neutralEyes, neutralMouth
	if t >= 0 {
		eyes += (raisedEyes[s.Head] - neutralEyes) * t
	} else {
		eyes += (squashedEyes - neutralEyes) * -t
		mouth += (squashedMouth - neutralMouth) * -t
	}
	return b.Y + b.H*eyes, b.Y + b.H*mouth
}

// blushY is the height of the cheek blush, between the eyes and the mouth.
func blushY(s Spec, b Layout) float64 {
	ey, my := facePos(s, b)
	return ey + (my-ey)*blushBetween
}

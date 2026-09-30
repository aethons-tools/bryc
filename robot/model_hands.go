package robot

import (
	"fmt"
	"math"

	"github.com/aethons-tools/bryc/robot/mesh"
)

// The 3D hands lie nearly flat in front of the face, palms down: hand
// space's v axis (toward the fingertips) points down handPitch from level,
// toward the viewer, and the backs of the hands face up and forward.
const (
	handPitch     = 35.0         // degrees below level the fingers point
	handWristZ    = faceZ + 50.0 // wrist distance in front of the head's centre
	palmThickness = 44.0
	palmBevel     = 18.0
)

func addHands(m *modelBuilder) {
	p := handPitch * math.Pi / 180
	// along is hand space's v direction in model space; across is u.
	along := mesh.Vec3{X: 0, Y: -math.Sin(p), Z: math.Cos(p)}
	for i, h := range hands(m.b) {
		wrist := upAt(h.wrist.X, h.wrist.Y, handWristZ)
		place := func(o mesh.Vec2) mesh.Vec3 {
			return wrist.Add(mesh.Vec3{X: o.X}).Add(along.Scale(o.Y))
		}
		// The palm is extruded in the XY plane with v pointing down (-Y),
		// then tipped forward so -Y becomes along.
		flat := make([]mesh.Vec2, len(h.palm))
		for k, o := range h.palm {
			flat[k] = mesh.Vec2{X: o.X, Y: -o.Y}
		}
		palm := mustExtrude(flat, palmThickness, palmBevel).RotateX(-(math.Pi/2 - p)).Translate(wrist)
		m.add(fmt.Sprintf("hand-%d", i), palm, m.mat.body)
		for j, d := range h.digits {
			name := fmt.Sprintf("finger-%d-%d", i, j)
			if j == 2 {
				name = fmt.Sprintf("thumb-%d", i)
			}
			root, tip := place(d.root), place(d.tip)
			r := d.r
			me := mesh.Tube([]mesh.Vec3{root, tip}, func(float64) float64 { return r })
			me.Append(mesh.Ellipsoid(r, r, r).Translate(tip))
			m.add(name, me, m.mat.body)
		}
	}
}

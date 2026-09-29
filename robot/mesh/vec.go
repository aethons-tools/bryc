// Package mesh builds triangle meshes: rounded extrusions of 2D outlines and
// simple solids. It knows nothing about robots.
package mesh

import "math"

// Vec2 is a 2D point or vector.
type Vec2 struct{ X, Y float64 }

// Vec3 is a 3D point or vector.
type Vec3 struct{ X, Y, Z float64 }

func (a Vec3) Add(b Vec3) Vec3      { return Vec3{a.X + b.X, a.Y + b.Y, a.Z + b.Z} }
func (a Vec3) Sub(b Vec3) Vec3      { return Vec3{a.X - b.X, a.Y - b.Y, a.Z - b.Z} }
func (a Vec3) Scale(k float64) Vec3 { return Vec3{a.X * k, a.Y * k, a.Z * k} }
func (a Vec3) Dot(b Vec3) float64   { return a.X*b.X + a.Y*b.Y + a.Z*b.Z }
func (a Vec3) Cross(b Vec3) Vec3 {
	return Vec3{a.Y*b.Z - a.Z*b.Y, a.Z*b.X - a.X*b.Z, a.X*b.Y - a.Y*b.X}
}
func (a Vec3) Len() float64 { return math.Sqrt(a.Dot(a)) }

// Norm returns a unit vector in a's direction, or a itself if it is zero.
func (a Vec3) Norm() Vec3 {
	if l := a.Len(); l > 0 {
		return a.Scale(1 / l)
	}
	return a
}

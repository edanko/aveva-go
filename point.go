package gen

import "math"

type Point struct {
	X float64
	Y float64
}

func (p Point) Vertice() []float64 {
	return []float64{p.X, p.Y, 0.0}
}

func (p Point) DistanceTo(other *Point) float64 {
	return math.Hypot(p.X-other.X, p.Y-other.Y)
}

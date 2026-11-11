package gen

import (
	"math"
)

func ToRad(deg float64) float64 {
	return deg * math.Pi / 180.0
}

func ToDeg(rad float64) float64 {
	return rad * 180.0 / math.Pi
}

func AddAngle(x, y, angle, length float64) (float64, float64) {
	a := ToRad(angle)
	newX := length*math.Cos(a) + x
	newY := length*math.Sin(a) + y

	return newX, newY
}

func Angle(start, end *Point) float64 {
	x := end.X - start.X
	y := end.Y - start.Y

	return math.Atan2(y, x)
}

func SegmentAngle(s *Segment) float64 {
	return Angle(s.Start, s.End)
}

func SegmentMidPoint(s *Segment) (x float64, y float64) {
	if s.Radius != 0.0 {
		x, y = ArcMidPoint(s.Start, s.End, s.Origin, s.Radius)
	} else {
		x, y = LineMidPoint(s.Start, s.End)
	}
	return
}

func MidPoint(start, end, center *Point, radius float64) (x float64, y float64) {
	if radius != 0.0 {
		x, y = ArcMidPoint(start, end, center, radius)
	} else {
		x, y = LineMidPoint(start, end)
	}
	return
}

func LineMidPoint(start, end *Point) (float64, float64) {
	x := start.X + ((end.X - start.X) / 2.0)
	y := start.Y + ((end.Y - start.Y) / 2.0)

	return x, y
}

func ArcMidPoint(start, end, center *Point, radius float64) (float64, float64) {
	a1 := Angle(center, start)
	a2 := Angle(center, end)

	theta := (a1 + a2) / 2.0

	newx := radius*math.Cos(theta) + center.X
	newy := radius*math.Sin(theta) + center.Y

	return newx, newy
}

package drawlib

import (
	"math"

	"golang.org/x/image/math/fixed"
)

type Vector struct {
	X, Y float64
}

func NewVector(x, y float64) *Vector {
	return &Vector{
		X: x,
		Y: y,
	}
}

func CreateQuadraticBezier(x0, y0, x1, y1, x2, y2 float64) []*Vector {
	l := (math.Hypot(x1-x0, y1-y0) +
		math.Hypot(x2-x1, y2-y1))
	n := int(l + 0.5)
	if n < 4 {
		n = 4
	}
	d := float64(n) - 1
	result := make([]*Vector, n)
	for i := 0; i < n; i++ {
		t := float64(i) / d
		u := 1 - t
		a := u * u
		b := 2 * u * t
		c := t * t
		result[i] = NewVector(a*x0+b*x1+c*x2, a*y0+b*y1+c*y2)
	}
	return result
}

func CreateCubicBezier(x0, y0, x1, y1, x2, y2, x3, y3 float64) []*Vector {
	l := (math.Hypot(x1-x0, y1-y0) +
		math.Hypot(x2-x1, y2-y1) +
		math.Hypot(x3-x2, y3-y2))
	n := int(l + 0.5)
	if n < 4 {
		n = 4
	}
	d := float64(n) - 1
	result := make([]*Vector, n)
	for i := 0; i < n; i++ {
		t := float64(i) / d
		u := 1 - t
		a := u * u * u
		b := 3 * u * u * t
		c := 3 * u * t * t
		d := t * t * t
		result[i] = NewVector(a*x0+b*x1+c*x2+d*x3, a*y0+b*y1+c*y2+d*y3)
	}
	return result
}

func (v *Vector) GetLength() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func (v *Vector) GetAngle() float64 {
	return math.Atan2(v.Y, v.X)
}

func (v *Vector) SetAngle(angle float64) {
	length := v.GetLength()
	v.X = math.Cos(angle) * length
	v.Y = math.Sin(angle) * length
}

func (v *Vector) SetLength(length float64) {
	angle := v.GetAngle()
	v.X = math.Cos(angle) * length
	v.Y = math.Sin(angle) * length
}

func (v *Vector) Add(v2 *Vector) *Vector {
	return NewVector(v.X+v2.X, v.Y+v2.Y)
}

func (v *Vector) Subtract(v2 *Vector) *Vector {
	return NewVector(v.X-v2.X, v.Y-v2.Y)
}

func (v *Vector) Multiply(v2 *Vector) *Vector {
	return NewVector(v.X*v2.X, v.Y*v2.Y)
}

func (v *Vector) Divide(v2 *Vector) *Vector {
	return NewVector(v.X/v2.X, v.Y/v2.Y)
}

func (v *Vector) AddTo(v2 *Vector) *Vector {
	v.X += v2.X
	v.Y += v2.Y
	return v
}

func (v *Vector) SubtractForm(v2 *Vector) *Vector {
	v.X -= v2.X
	v.Y -= v2.Y
	return v
}

func (v *Vector) MultiplyBy(v2 *Vector) *Vector {
	v.X *= v2.X
	v.Y *= v2.Y
	return v
}

func (v *Vector) DivideBy(v2 *Vector) *Vector {
	v.X /= v2.X
	v.Y /= v2.Y
	return v
}

func (v *Vector) Dot(v2 *Vector) float64 {
	return v.X*v2.X + v.Y*v2.Y
}

func (v *Vector) Distance(v2 *Vector) float64 {
	dx := v.X - v2.X
	dy := v.Y - v2.Y
	return math.Sqrt(dx*dx + dy*dy)
}

func (v *Vector) Unit() float64 {
	return 1.0 / math.Sqrt(v.X*v.X+v.Y*v.Y)
}

func (v *Vector) Copy() *Vector {
	return NewVector(v.X, v.Y)
}

func (v *Vector) Negative() {
	v.X = -v.X
	v.Y = -v.Y
}

func (v *Vector) PerpendicularWith(v2 *Vector) {
	v.X = -v2.Y
	v.Y = v2.X
}

func (v *Vector) Perpendicular() {
	temp := v.X
	v.X = -v.Y
	v.Y = temp
}

func (v *Vector) Normalize() {
	u := v.Unit()
	v.X = v.X * u
	v.Y = v.Y * u
}

func (v *Vector) Fixed() fixed.Point26_6 {
	return fixed.Point26_6{
		X: fixed.I(int(v.X)),
		Y: fixed.I(int(v.Y)),
	}
}

func (v *Vector) Interpolate(v2 *Vector, t float64) *Vector {
	x := v.X + (v2.X-v.X)*t
	y := v.Y + (v2.Y-v.Y)*t
	return NewVector(x, y)
}

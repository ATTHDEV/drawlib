package drawlib

import "math"

type Matrix struct {
	XX, YX, XY, YY, X0, Y0 float64
}

func NewMatrix() *Matrix {
	return Identity()
}

func Identity() *Matrix {
	return &Matrix{
		1, 0,
		0, 1,
		0, 0,
	}
}

func Translate(x, y float64) *Matrix {
	return &Matrix{
		1, 0,
		0, 1,
		x, y,
	}
}

func Scale(x, y float64) *Matrix {
	return &Matrix{
		x, 0,
		0, y,
		0, 0,
	}
}

func Rotate(angle float64) *Matrix {
	c := math.Cos(angle)
	s := math.Sin(angle)
	return &Matrix{
		c, s,
		-s, c,
		0, 0,
	}
}

func Shear(x, y float64) *Matrix {
	return &Matrix{
		1, y,
		x, 1,
		0, 0,
	}
}

func (m1 Matrix) Multiply(m2 Matrix) *Matrix {
	return &Matrix{
		m1.XX*m2.XX + m1.YX*m2.XY,
		m1.XX*m2.YX + m1.YX*m2.YY,
		m1.XY*m2.XX + m1.YY*m2.XY,
		m1.XY*m2.YX + m1.YY*m2.YY,
		m1.X0*m2.XX + m1.Y0*m2.XY + m2.X0,
		m1.X0*m2.YX + m1.Y0*m2.YY + m2.Y0,
	}
}

func (m Matrix) TransformVector(x, y float64) (tx, ty float64) {
	return m.XX*x + m.XY*y, m.YX*x + m.YY*y
}

func (m Matrix) TransformPoint(x, y float64) (tx, ty float64) {
	return m.XX*x + m.XY*y + m.X0, m.YX*x + m.YY*y + m.Y0
}

func (m Matrix) Translate(x, y float64) *Matrix {
	return Translate(x, y).Multiply(m)
}

func (m Matrix) Scale(x, y float64) *Matrix {
	return Scale(x, y).Multiply(m)
}

func (m Matrix) Rotate(angle float64) *Matrix {
	return Rotate(angle).Multiply(m)
}

func (m Matrix) Shear(x, y float64) *Matrix {
	return Shear(x, y).Multiply(m)
}

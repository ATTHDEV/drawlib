package drawlib

import (
	"github.com/golang/freetype/raster"
	"golang.org/x/image/math/fixed"
)

func flattenPath(p raster.Path) [][]*Vector {
	var result [][]*Vector
	var path []*Vector
	var cx, cy float64
	for i := 0; i < len(p); {
		switch p[i] {
		case 0:
			if len(path) > 0 {
				result = append(result, path)
				path = nil
			}
			x := Unfix(p[i+1])
			y := Unfix(p[i+2])
			path = append(path, NewVector(x, y))
			cx, cy = x, y
			i += 4
		case 1:
			x := Unfix(p[i+1])
			y := Unfix(p[i+2])
			path = append(path, NewVector(x, y))
			cx, cy = x, y
			i += 4
		case 2:
			x1 := Unfix(p[i+1])
			y1 := Unfix(p[i+2])
			x2 := Unfix(p[i+3])
			y2 := Unfix(p[i+4])
			points := CreateQuadraticBezier(cx, cy, x1, y1, x2, y2)
			path = append(path, points...)
			cx, cy = x2, y2
			i += 6
		case 3:
			x1 := Unfix(p[i+1])
			y1 := Unfix(p[i+2])
			x2 := Unfix(p[i+3])
			y2 := Unfix(p[i+4])
			x3 := Unfix(p[i+5])
			y3 := Unfix(p[i+6])
			points := CreateCubicBezier(cx, cy, x1, y1, x2, y2, x3, y3)
			path = append(path, points...)
			cx, cy = x3, y3
			i += 8
		default:
			panic("bad path")
		}
	}
	if len(path) > 0 {
		result = append(result, path)
	}
	return result
}

func dashPath(paths [][]*Vector, dashes []float64) [][]*Vector {
	var result [][]*Vector
	if len(dashes) == 0 {
		return paths
	}
	if len(dashes) == 1 {
		dashes = append(dashes, dashes[0])
	}
	for _, path := range paths {
		if len(path) < 2 {
			continue
		}
		previous := path[0]
		pathIndex := 1
		dashIndex := 0
		segmentLength := 0.0
		var segment []*Vector
		segment = append(segment, previous)
		for pathIndex < len(path) {
			dashLength := dashes[dashIndex]
			point := path[pathIndex]
			d := previous.Distance(point)
			maxd := dashLength - segmentLength
			if d > maxd {
				t := maxd / d
				p := previous.Interpolate(point, t)
				segment = append(segment, p)
				if dashIndex%2 == 0 && len(segment) > 1 {
					result = append(result, segment)
				}
				segment = nil
				segment = append(segment, p)
				segmentLength = 0
				previous = p
				dashIndex = (dashIndex + 1) % len(dashes)
			} else {
				segment = append(segment, point)
				previous = point
				segmentLength += d
				pathIndex++
			}
		}
		if dashIndex%2 == 0 && len(segment) > 1 {
			result = append(result, segment)
		}
	}
	return result
}

func rasterPath(paths [][]*Vector) raster.Path {
	var result raster.Path
	for _, path := range paths {
		var previous fixed.Point26_6
		for i, point := range path {
			f := point.Fixed()
			if i == 0 {
				result.Start(f)
			} else {
				dx := f.X - previous.X
				dy := f.Y - previous.Y
				if dx < 0 {
					dx = -dx
				}
				if dy < 0 {
					dy = -dy
				}
				if dx+dy > 8 {
					result.Add1(f)
				}
			}
			previous = f
		}
	}
	return result
}

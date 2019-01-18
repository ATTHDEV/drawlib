package drawlib

import (
	"errors"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"strings"
	"unicode"

	"github.com/golang/freetype/raster"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/f64"
)

type LineCap int
type LineJoin int
type FillRule int
type Align int

const (
	LineCapRound LineCap = iota
	LineCapButt
	LineCapSquare

	LineJoinRound LineJoin = iota
	LineJoinBevel

	FillRuleWinding FillRule = iota
	FillRuleEvenOdd

	AlignLeft Align = iota
	AlignCenter
	AlignRight
)

var (
	defaultFillStyle   = NewSolidPattern(color.White)
	defaultStrokeStyle = NewSolidPattern(color.Black)
)

type Canvas struct {
	stack         []*Canvas
	width         int
	height        int
	rasterizer    *raster.Rasterizer
	im            *image.RGBA
	mask          *image.Alpha
	color         color.Color
	clearSrc      *image.Uniform
	fillPattern   Pattern
	strokePattern Pattern
	strokePath    raster.Path
	fillPath      raster.Path
	start         *Vector
	current       *Vector
	hasCurrent    bool
	hasBegin      bool
	dashes        []float64
	vertexts      []*Vector
	lineWidth     float64
	lineCap       LineCap
	lineJoin      LineJoin
	fillRule      FillRule
	fontFace      font.Face
	fontHeight    float64
	matrix        *Matrix
}

func NewCanvas(width, height int) *Canvas {
	return NewCanvasForRGBA(image.NewRGBA(image.Rect(0, 0, width, height)))
}

func NewCanvasForImage(im image.Image) *Canvas {
	return NewCanvasForRGBA(ImageToRGBA(im))
}

func NewCanvasForRGBA(im *image.RGBA) *Canvas {
	w := im.Bounds().Size().X
	h := im.Bounds().Size().Y
	return &Canvas{
		width:         w,
		height:        h,
		rasterizer:    raster.NewRasterizer(w, h),
		im:            im,
		color:         color.Transparent,
		clearSrc:      image.White,
		fillPattern:   defaultFillStyle,
		strokePattern: defaultStrokeStyle,
		lineWidth:     1,
		fillRule:      FillRuleWinding,
		fontFace:      basicfont.Face7x13,
		fontHeight:    13,
		matrix:        Identity(),
	}
}

func (c *Canvas) Image() image.Image {
	return c.im
}

func (c *Canvas) Width() int {
	return c.width
}

func (c *Canvas) Height() int {
	return c.height
}

func (c *Canvas) SavePNG(path string) error {
	return SavePNG(path, c.im)
}

func (c *Canvas) EncodePNG(w io.Writer) error {
	return png.Encode(w, c.im)
}

func (c *Canvas) SetDash(dashes ...float64) *Canvas {
	c.dashes = dashes
	return c
}

func (c *Canvas) SetLineWidth(lineWidth float64) *Canvas {
	c.lineWidth = lineWidth
	return c
}

func (c *Canvas) SetLineCap(lineCap LineCap) *Canvas {
	c.lineCap = lineCap
	return c
}

func (c *Canvas) SetLineCapRound() *Canvas {
	c.lineCap = LineCapRound
	return c
}

func (c *Canvas) SetLineCapButt() *Canvas {
	c.lineCap = LineCapButt
	return c
}

func (c *Canvas) SetLineCapSquare() *Canvas {
	c.lineCap = LineCapSquare
	return c
}

func (c *Canvas) SetLineJoin(lineJoin LineJoin) *Canvas {
	c.lineJoin = lineJoin
	return c
}

func (c *Canvas) SetLineJoinRound() *Canvas {
	c.lineJoin = LineJoinRound
	return c
}

func (c *Canvas) SetLineJoinBevel() *Canvas {
	c.lineJoin = LineJoinBevel
	return c
}

func (c *Canvas) SetFillRule(fillRule FillRule) *Canvas {
	c.fillRule = fillRule
	return c
}

func (c *Canvas) SetFillRuleWinding() *Canvas {
	c.fillRule = FillRuleWinding
	return c
}

func (c *Canvas) SetFillRuleEvenOdd() *Canvas {
	c.fillRule = FillRuleEvenOdd
	return c
}

func (c *Canvas) setFillAndStrokeColor(color color.Color) *Canvas {
	c.color = color
	c.fillPattern = NewSolidPattern(color)
	c.strokePattern = NewSolidPattern(color)
	return c
}

func (c *Canvas) SetFillStyle(pattern Pattern) *Canvas {
	if fillStyle, ok := pattern.(*solidPattern); ok {
		c.color = fillStyle.color
	}
	c.fillPattern = pattern
	return c
}

func (c *Canvas) SetStrokeStyle(pattern Pattern) *Canvas {
	c.strokePattern = pattern
	return c
}

func (c *Canvas) SetColor(color color.Color) *Canvas {
	c.setFillAndStrokeColor(color)
	return c
}

func (c *Canvas) SetHexColor(x string) *Canvas {
	r, g, b, a := HexToRGBA(x)
	c.SetRGBA255(r, g, b, a)
	return c
}

func (c *Canvas) SetRGBA255(r, g, b, a int) *Canvas {
	c.color = color.NRGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
	c.setFillAndStrokeColor(c.color)
	return c
}

func (c *Canvas) SetRGB255(r, g, b int) *Canvas {
	c.SetRGBA255(r, g, b, 255)
	return c
}

func (c *Canvas) SetRGBA(r, g, b, a float64) *Canvas {
	c.color = color.NRGBA{
		uint8(r * 255),
		uint8(g * 255),
		uint8(b * 255),
		uint8(a * 255),
	}
	c.setFillAndStrokeColor(c.color)
	return c
}

func (c *Canvas) SetRGB(r, g, b float64) *Canvas {
	c.SetRGBA(r, g, b, 1)
	return c
}

func (c *Canvas) MoveTo(x, y float64) *Canvas {
	if c.hasCurrent {
		c.fillPath.Add1(c.start.Fixed())
	}
	x, y = c.TransformPoint(x, y)
	v := NewVector(x, y)
	c.strokePath.Start(v.Fixed())
	c.fillPath.Start(v.Fixed())
	c.start = v
	c.current = v
	c.hasCurrent = true
	return c
}

func (c *Canvas) LineTo(x, y float64) *Canvas {
	if !c.hasCurrent {
		c.MoveTo(x, y)
	} else {
		x, y = c.TransformPoint(x, y)
		p := NewVector(x, y)
		c.strokePath.Add1(p.Fixed())
		c.fillPath.Add1(p.Fixed())
		c.current = p
	}
	return c
}

func (c *Canvas) BeginShape() *Canvas {
	c.vertexts = []*Vector{}
	c.hasBegin = true
	return c
}

func (c *Canvas) Vertex(x, y float64) *Canvas {
	if c.hasBegin {
		c.vertexts = append(c.vertexts, NewVector(x, y))
	}
	return c
}

func (c *Canvas) EndShape(isClose ...bool) *Canvas {
	if c.hasBegin {
		c.hasBegin = false
		c.MoveTo(c.vertexts[0].X, c.vertexts[0].Y)
		for i := 1; i < len(c.vertexts); i++ {
			c.LineTo(c.vertexts[i].X, c.vertexts[i].Y)
		}

		if len(isClose) == 1 {
			if isClose[0] == true {
				c.LineTo(c.vertexts[0].X, c.vertexts[0].Y)
			}
		}
	}
	return c
}

func (c *Canvas) QuadraticTo(x1, y1, x2, y2 float64) *Canvas {
	if !c.hasCurrent {
		c.MoveTo(x1, y1)
	}
	x1, y1 = c.TransformPoint(x1, y1)
	x2, y2 = c.TransformPoint(x2, y2)
	v1 := NewVector(x1, y1)
	v2 := NewVector(x2, y2)
	c.strokePath.Add2(v1.Fixed(), v2.Fixed())
	c.fillPath.Add2(v1.Fixed(), v2.Fixed())
	c.current = v2
	return c
}

func (c *Canvas) CubicTo(x1, y1, x2, y2, x3, y3 float64) *Canvas {
	if !c.hasCurrent {
		c.MoveTo(x1, y1)
	}
	x0, y0 := c.current.X, c.current.Y
	x1, y1 = c.TransformPoint(x1, y1)
	x2, y2 = c.TransformPoint(x2, y2)
	x3, y3 = c.TransformPoint(x3, y3)

	points := CreateCubicBezier(x0, y0, x1, y1, x2, y2, x3, y3)
	previous := c.current.Fixed()
	for _, p := range points[1:] {
		f := p.Fixed()
		if f == previous {
			continue
		}
		previous = f
		c.strokePath.Add1(f)
		c.fillPath.Add1(f)
		c.current = p
	}
	return c
}

func (c *Canvas) ClosePath() *Canvas {
	if c.hasCurrent {
		c.strokePath.Add1(c.start.Fixed())
		c.fillPath.Add1(c.start.Fixed())
		c.current = c.start
	}
	return c
}

func (c *Canvas) ClearPath() *Canvas {
	c.strokePath.Clear()
	c.fillPath.Clear()
	c.hasCurrent = false
	return c
}

func (c *Canvas) NewSubPath() *Canvas {
	if c.hasCurrent {
		c.fillPath.Add1(c.start.Fixed())
	}
	c.hasCurrent = false
	return c
}

func (c *Canvas) capper() raster.Capper {
	switch c.lineCap {
	case LineCapButt:
		return raster.ButtCapper
	case LineCapRound:
		return raster.RoundCapper
	case LineCapSquare:
		return raster.SquareCapper
	}
	return nil
}

func (c *Canvas) joiner() raster.Joiner {
	switch c.lineJoin {
	case LineJoinBevel:
		return raster.BevelJoiner
	case LineJoinRound:
		return raster.RoundJoiner
	}
	return nil
}

func (c *Canvas) stroke(painter raster.Painter) *Canvas {
	path := c.strokePath
	if len(c.dashes) > 0 {
		path = rasterPath(dashPath(flattenPath(path), c.dashes))
	} else {
		path = rasterPath(flattenPath(path))
	}
	r := c.rasterizer
	r.UseNonZeroWinding = true
	r.Clear()
	r.AddStroke(path, Fix(c.lineWidth), c.capper(), c.joiner())
	r.Rasterize(painter)
	return c
}

func (c *Canvas) fill(painter raster.Painter) *Canvas {
	path := c.fillPath
	if c.hasCurrent {
		path = make(raster.Path, len(c.fillPath))
		copy(path, c.fillPath)
		path.Add1(c.start.Fixed())
	}
	r := c.rasterizer
	r.UseNonZeroWinding = c.fillRule == FillRuleWinding
	r.Clear()
	r.AddPath(path)
	r.Rasterize(painter)
	return c
}

func (c *Canvas) StrokePreserve() *Canvas {
	var painter raster.Painter
	if c.mask == nil {
		if pattern, ok := c.strokePattern.(*solidPattern); ok {
			p := raster.NewRGBAPainter(c.im)
			p.SetColor(pattern.color)
			painter = p
		}
	}
	if painter == nil {
		painter = newPatternPainter(c.im, c.mask, c.strokePattern)
	}
	c.stroke(painter)
	return c
}

func (c *Canvas) Stroke() {
	c.StrokePreserve()
	c.ClearPath()
}

func (c *Canvas) StrokeRGB(r, g, b float64) {
	c.SetRGB(r, g, b)
	c.Stroke()
}

func (c *Canvas) StrokeRGBA(r, g, b, a float64) {
	c.SetRGBA(r, g, b, a)
	c.Stroke()
}

func (c *Canvas) StrokeRGB255(r, g, b int) {
	c.SetRGB255(r, g, b)
	c.Stroke()
}

func (c *Canvas) StrokeRGBA255(r, g, b, a int) {
	c.SetRGBA255(r, g, b, a)
	c.Stroke()
}

func (c *Canvas) FillPreserve() *Canvas {
	var painter raster.Painter
	if c.mask == nil {
		if pattern, ok := c.fillPattern.(*solidPattern); ok {
			p := raster.NewRGBAPainter(c.im)
			p.SetColor(pattern.color)
			painter = p
		}
	}
	if painter == nil {
		painter = newPatternPainter(c.im, c.mask, c.fillPattern)
	}
	c.fill(painter)
	return c
}

func (c *Canvas) Fill() {
	c.FillPreserve()
	c.ClearPath()
}

func (c *Canvas) FillRGB(r, g, b float64) {
	c.SetRGB(r, g, b)
	c.Fill()
}

func (c *Canvas) FillRGBA(r, g, b, a float64) {
	c.SetRGBA(r, g, b, a)
	c.Fill()
}

func (c *Canvas) FillRGB255(r, g, b int) {
	c.SetRGB255(r, g, b)
	c.Fill()
}

func (c *Canvas) FillRGBA255(r, g, b, a int) {
	c.SetRGBA255(r, g, b, a)
	c.Fill()
}

func (c *Canvas) ClipPreserve() *Canvas {
	clip := image.NewAlpha(image.Rect(0, 0, c.width, c.height))
	painter := raster.NewAlphaOverPainter(clip)
	c.fill(painter)
	if c.mask == nil {
		c.mask = clip
	} else {
		mask := image.NewAlpha(image.Rect(0, 0, c.width, c.height))
		draw.DrawMask(mask, mask.Bounds(), clip, image.ZP, c.mask, image.ZP, draw.Over)
		c.mask = mask
	}
	return c
}

func (c *Canvas) SetMask(mask *image.Alpha) error {
	if mask.Bounds().Size() != c.im.Bounds().Size() {
		return errors.New("mask size must match context size")
	}
	c.mask = mask
	return nil
}

func (c *Canvas) AsMask() *image.Alpha {
	mask := image.NewAlpha(c.im.Bounds())
	draw.Draw(mask, c.im.Bounds(), c.im, image.ZP, draw.Src)
	return mask
}

func (c *Canvas) InvertMask() *Canvas {
	if c.mask == nil {
		c.mask = image.NewAlpha(c.im.Bounds())
	} else {
		for i, a := range c.mask.Pix {
			c.mask.Pix[i] = 255 - a
		}
	}
	return c
}

func (c *Canvas) Clip() *Canvas {
	c.ClipPreserve()
	c.ClearPath()
	return c
}

func (c *Canvas) ResetClip() *Canvas {
	c.mask = nil
	return c
}

// arg 1 for gray color,
// arg 3 for rgb color,
// arg 4 for rgba color
func (c *Canvas) SetClearColor(i ...interface{}) *Canvas {
	var col color.Color
	switch len(i) {
	case 1:
		v := uint8(i[0].(int))
		col = color.RGBA{v, v, v, 255}
	case 3:
		if v, ok := i[0].(int); ok {
			col = color.RGBA{uint8(v), uint8(i[1].(int)), uint8(i[2].(int)), 255}
		} else {
			col = color.RGBA{uint8(i[0].(float64)), uint8(i[1].(float64)), uint8(i[2].(float64)), 255}
		}
	case 4:
		if v, ok := i[0].(int); ok {
			col = color.RGBA{uint8(v), uint8(i[1].(int)), uint8(i[2].(int)), uint8(i[3].(int))}
		} else {
			col = color.RGBA{uint8(i[0].(float64)), uint8(i[1].(float64)), uint8(i[2].(float64)), uint8(i[3].(float64))}
		}
	}
	c.clearSrc = image.NewUniform(col)
	return c
}

// arg 1 for gray color,
// arg 3 for rgb color,
// arg 4 for rgba color
func (c *Canvas) Background(i ...interface{}) *Canvas {
	c.SetClearColor(i...)
	c.Clear()
	return c
}

func (c *Canvas) Clear() *Canvas {
	draw.Draw(c.im, c.im.Bounds(), c.clearSrc, image.ZP, draw.Src)
	return c
}

func (c *Canvas) SetPixel(x, y int) *Canvas {
	c.im.Set(x, y, c.color)
	return c
}

func (c *Canvas) DrawPoint(x, y, r float64) *Canvas {
	c.Push()
	tx, ty := c.TransformPoint(x, y)
	c.Identity()
	c.DrawCircle(tx, ty, r)
	c.Pop()
	return c
}

func (c *Canvas) DrawLine(x1, y1, x2, y2 float64) *Canvas {
	c.MoveTo(x1, y1)
	c.LineTo(x2, y2)
	return c
}

func (c *Canvas) DrawRectangle(x, y, w, h float64) *Canvas {
	c.NewSubPath()
	c.MoveTo(x, y)
	c.LineTo(x+w, y)
	c.LineTo(x+w, y+h)
	c.LineTo(x, y+h)
	c.ClosePath()
	return c
}

func (c *Canvas) DrawRoundedRectangle(x, y, w, h, r float64) *Canvas {
	x0, x1, x2, x3 := x, x+r, x+w-r, x+w
	y0, y1, y2, y3 := y, y+r, y+h-r, y+h
	c.NewSubPath()
	c.MoveTo(x1, y0)
	c.LineTo(x2, y0)
	c.DrawArc(x2, y1, r, ToRadians(270), ToRadians(360))
	c.LineTo(x3, y2)
	c.DrawArc(x2, y2, r, ToRadians(0), ToRadians(90))
	c.LineTo(x1, y3)
	c.DrawArc(x1, y2, r, ToRadians(90), ToRadians(180))
	c.LineTo(x0, y1)
	c.DrawArc(x1, y1, r, ToRadians(180), ToRadians(270))
	c.ClosePath()
	return c
}

func (c *Canvas) DrawEllipticalArc(x, y, rx, ry, angle1, angle2 float64) *Canvas {
	const n = 16
	for i := 0; i < n; i++ {
		p1 := float64(i+0) / n
		p2 := float64(i+1) / n
		a1 := angle1 + (angle2-angle1)*p1
		a2 := angle1 + (angle2-angle1)*p2
		x0 := x + rx*math.Cos(a1)
		y0 := y + ry*math.Sin(a1)
		x1 := x + rx*math.Cos(a1+(a2-a1)/2)
		y1 := y + ry*math.Sin(a1+(a2-a1)/2)
		x2 := x + rx*math.Cos(a2)
		y2 := y + ry*math.Sin(a2)
		cx := 2*x1 - x0/2 - x2/2
		cy := 2*y1 - y0/2 - y2/2
		if i == 0 {
			if c.hasCurrent {
				c.LineTo(x0, y0)
			} else {
				c.MoveTo(x0, y0)
			}
		}
		c.QuadraticTo(cx, cy, x2, y2)
	}
	return c
}

func (c *Canvas) DrawEllipse(x, y, rx, ry float64) *Canvas {
	c.NewSubPath()
	c.DrawEllipticalArc(x, y, rx, ry, 0, 2*math.Pi)
	c.ClosePath()
	return c
}

func (c *Canvas) DrawArc(x, y, r, angle1, angle2 float64) *Canvas {
	c.DrawEllipticalArc(x, y, r, r, angle1, angle2)
	return c
}

func (c *Canvas) DrawCircle(x, y, r float64) *Canvas {
	c.NewSubPath()
	c.DrawEllipticalArc(x, y, r, r, 0, 2*math.Pi)
	c.ClosePath()
	return c
}

func (c *Canvas) DrawRegularPolygon(n int, x, y, r, rotation float64) *Canvas {
	angle := 2 * math.Pi / float64(n)
	rotation -= math.Pi / 2
	if n%2 == 0 {
		rotation += angle / 2
	}
	c.NewSubPath()
	for i := 0; i < n; i++ {
		a := rotation + angle*float64(i)
		c.LineTo(x+r*math.Cos(a), y+r*math.Sin(a))
	}
	c.ClosePath()
	return c
}

func (c *Canvas) DrawImage(im image.Image, x, y int) *Canvas {
	c.DrawImageAnchored(im, x, y, 0, 0)
	return c
}

func (c *Canvas) DrawImageAnchored(im image.Image, x, y int, ax, ay float64) *Canvas {
	s := im.Bounds().Size()
	x -= int(ax * float64(s.X))
	y -= int(ay * float64(s.Y))
	transformer := draw.BiLinear
	fx, fy := float64(x), float64(y)
	m := c.matrix.Translate(fx, fy)
	s2d := f64.Aff3{m.XX, m.XY, m.X0, m.YX, m.YY, m.Y0}
	if c.mask == nil {
		transformer.Transform(c.im, s2d, im, im.Bounds(), draw.Over, nil)
	} else {
		transformer.Transform(c.im, s2d, im, im.Bounds(), draw.Over, &draw.Options{
			DstMask:  c.mask,
			DstMaskP: image.ZP,
		})
	}
	return c
}

func (c *Canvas) SetFontFace(fontFace font.Face) *Canvas {
	c.fontFace = fontFace
	c.fontHeight = float64(fontFace.Metrics().Height) / 64
	return c
}

func (c *Canvas) LoadFontFace(path string, points float64) error {
	face, err := LoadFontFace(path, points)
	if err == nil {
		c.fontFace = face
		c.fontHeight = points * 72 / 96
	}
	return err
}

func (c *Canvas) FontHeight() float64 {
	return c.fontHeight
}

func (c *Canvas) drawString(im *image.RGBA, s string, x, y float64) {
	d := &font.Drawer{
		Dst:  im,
		Src:  image.NewUniform(c.color),
		Face: c.fontFace,
		Dot:  Fixp(x, y),
	}
	prevC := rune(-1)
	for _, r := range s {
		if prevC >= 0 {
			d.Dot.X += d.Face.Kern(prevC, r)
		}
		dr, mask, maskp, advance, ok := d.Face.Glyph(d.Dot, r)
		if !ok {
			continue
		}
		sr := dr.Sub(dr.Min)
		transformer := draw.BiLinear
		fx, fy := float64(dr.Min.X), float64(dr.Min.Y)
		m := c.matrix.Translate(fx, fy)
		s2d := f64.Aff3{m.XX, m.XY, m.X0, m.YX, m.YY, m.Y0}
		transformer.Transform(d.Dst, s2d, d.Src, sr, draw.Over, &draw.Options{
			SrcMask:  mask,
			SrcMaskP: maskp,
		})
		d.Dot.X += advance
		prevC = r
	}
}

func (c *Canvas) DrawString(s string, x, y float64) *Canvas {
	c.DrawStringAnchored(s, x, y, 0, 0)
	return c
}

func (c *Canvas) DrawStringAnchored(s string, x, y, ax, ay float64) *Canvas {
	w, h := c.MeasureString(s)
	x -= ax * w
	y += ay * h
	if c.mask == nil {
		c.drawString(c.im, s, x, y)
	} else {
		im := image.NewRGBA(image.Rect(0, 0, c.width, c.height))
		c.drawString(im, s, x, y)
		draw.DrawMask(c.im, c.im.Bounds(), im, image.ZP, c.mask, image.ZP, draw.Over)
	}
	return c
}

func (c *Canvas) DrawStringWrapped(s string, x, y, ax, ay, width, lineSpacing float64, align Align) *Canvas {
	lines := c.WordWrap(s, width)

	h := float64(len(lines)) * c.fontHeight * lineSpacing
	h -= (lineSpacing - 1) * c.fontHeight

	x -= ax * width
	y -= ay * h
	switch align {
	case AlignLeft:
		ax = 0
	case AlignCenter:
		ax = 0.5
		x += width / 2
	case AlignRight:
		ax = 1
		x += width
	}
	ay = 1
	for _, line := range lines {
		c.DrawStringAnchored(line, x, y, ax, ay)
		y += c.fontHeight * lineSpacing
	}
	return c
}

func (c *Canvas) MeasureMultilineString(s string, lineSpacing float64) (width, height float64) {
	lines := strings.Split(s, "\n")

	height = float64(len(lines)) * c.fontHeight * lineSpacing
	height -= (lineSpacing - 1) * c.fontHeight

	d := &font.Drawer{
		Face: c.fontFace,
	}

	for _, line := range lines {
		adv := d.MeasureString(line)
		currentWidth := float64(adv >> 6)
		if currentWidth > width {
			width = currentWidth
		}
	}
	return width, height
}

func (c *Canvas) MeasureString(s string) (w, h float64) {
	d := &font.Drawer{
		Face: c.fontFace,
	}
	a := d.MeasureString(s)
	return float64(a >> 6), c.fontHeight
}

func (c *Canvas) WordWrap(s string, width float64) []string {
	var result []string
	for _, line := range strings.Split(s, "\n") {

		var fields []string
		pi := 0
		ps := false
		for i, c := range line {
			s := unicode.IsSpace(c)
			if s != ps && i > 0 {
				result = append(result, line[pi:i])
				pi = i
			}
			ps = s
		}
		fields = append(result, line[pi:])

		if len(fields)%2 == 1 {
			fields = append(fields, "")
		}
		x := ""
		for i := 0; i < len(fields); i += 2 {
			w, _ := c.MeasureString(x + fields[i])
			if w > width {
				if x == "" {
					result = append(result, fields[i])
					x = ""
					continue
				} else {
					result = append(result, x)
					x = ""
				}
			}
			x += fields[i] + fields[i+1]
		}
		if x != "" {
			result = append(result, x)
		}
	}
	for i, line := range result {
		result[i] = strings.TrimSpace(line)
	}
	return result
}

func (c *Canvas) Identity() *Canvas {
	c.matrix = Identity()
	return c
}

func (c *Canvas) Translate(x, y float64) *Canvas {
	c.matrix = c.matrix.Translate(x, y)
	return c
}

func (c *Canvas) Scale(x, y float64) *Canvas {
	c.matrix = c.matrix.Scale(x, y)
	return c
}

func (c *Canvas) ScaleAbout(sx, sy, x, y float64) *Canvas {
	c.Translate(x, y)
	c.Scale(sx, sy)
	c.Translate(-x, -y)
	return c
}

func (c *Canvas) Rotate(angle float64) *Canvas {
	c.matrix = c.matrix.Rotate(angle)
	return c
}

func (c *Canvas) RotateAbout(angle, x, y float64) *Canvas {
	c.Translate(x, y)
	c.Rotate(angle)
	c.Translate(-x, -y)
	return c
}

func (c *Canvas) Shear(x, y float64) *Canvas {
	c.matrix = c.matrix.Shear(x, y)
	return c
}

func (c *Canvas) ShearAbout(sx, sy, x, y float64) *Canvas {
	c.Translate(x, y)
	c.Shear(sx, sy)
	c.Translate(-x, -y)
	return c
}

func (c *Canvas) TransformPoint(x, y float64) (tx, ty float64) {
	return c.matrix.TransformPoint(x, y)
}

func (c *Canvas) InvertY() *Canvas {
	c.Translate(0, float64(c.height))
	c.Scale(1, -1)
	return c
}

func (c *Canvas) Push() *Canvas {
	c.stack = append(c.stack, c)
	return c
}

func (c *Canvas) Pop() *Canvas {
	before := *c
	s := c.stack
	x, s := s[len(s)-1], s[:len(s)-1]
	*c = *x
	c.mask = before.mask
	c.strokePath = before.strokePath
	c.fillPath = before.fillPath
	c.start = before.start
	c.current = before.current
	c.hasCurrent = before.hasCurrent
	return c
}

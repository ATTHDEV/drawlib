package main

import (
	"math/rand"
	"time"

	"github.com/ATTHDEV/drawlib"
)

type Object struct {
	s *drawlib.Vector
	v *drawlib.Vector
	r float64
	a int
}

func main() {
	d := drawlib.New(drawlib.Option().Title("Hello").Dimension(600, 600))

	var (
		sx, sy = d.Width(), d.Height()
		dc     = d.Canvas
		start  = drawlib.NewVector(sx/2, sy-100)
		items  = []*Object{}
	)

	go func() {
		// spawn loop
		for {
			time.Sleep(70 * time.Millisecond)
			items = append(items, &Object{
				r: 10 + rand.Float64()*15,
				s: start.Copy(),
				v: drawlib.NewVector(-2+rand.Float64()*(3 - -2), -3+rand.Float64()*-7),
				a: 255,
			})
		}
	}()

	d.RenderLoop(func(t float64) {
		dc.Background(0)
		// draw smoke
		for _, item := range items {
			s := item.s
			dc.Push()
			dc.DrawCircle(s.X, s.Y, item.r)
			dc.FillRGBA255(255, 255, 255, item.a)
			dc.Pop()
		}
		// update smoke
		for i, item := range items {
			item.a -= 3
			item.s.AddTo(item.v)
			if item.a <= 0 {
				items = append(items[:i], items[i+1:]...)
			}
		}
	})
	d.Start()
}

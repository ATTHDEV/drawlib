// simple snake game
package main

import (
	"fmt"
	"math/rand"
	"time"

	"golang.org/x/mobile/event/key"

	"github.com/ATTHDEV/drawlib"
)

func main() {
	const (
		w, h, size = 600, 600, 20
	)
	var (
		s        = 0
		gameOver = false
		speed    = float64(size + 1)
	)

	d := drawlib.New(drawlib.Option().Title("Hello").Dimension(w, h))
	v := drawlib.NewVector(0, size+1)
	body, foods := []*drawlib.Vector{}, []*drawlib.Vector{}
	for i := 5.0; i > 0; i-- {
		body = append(body, drawlib.NewVector(5*speed, i*speed))
	}

	dc := d.Canvas
	dc.LoadFontFace(drawlib.TAHOMA, 32)

	// spawn food
	go func() {
		for {
			foods = append(foods, drawlib.NewVector(float64(rand.Intn(w/size))*speed, float64(rand.Intn(h/size))*speed))
			time.Sleep(2 * time.Second)
		}
	}()

	d.RenderLoop(func(dt float64) {
		if s < 10 {
			time.Sleep(100 * time.Millisecond)
		} else {
			time.Sleep(40 * time.Millisecond)
		}
		dc.Background(0)
		// draw foods
		for _, f := range foods {
			dc.DrawCircle(f.X+10, f.Y+10, 7.5)
		}
		dc.FillRGB255(128, 128, 128)

		// draw snake
		for _, b := range body {
			dc.DrawRectangle(b.X, b.Y, size, size)
		}
		dc.FillRGB255(255, 255, 255)

		// draw score
		dc.DrawString(fmt.Sprintf("score : %d", s), 10, 32)
		if !gameOver {
			// check eat food (simple method)
			for i, f := range foods {
				if body[0].X == f.X && body[0].Y == f.Y {
					foods = append(foods[:i], foods[i+1:]...)
					body = append(body, body[len(body)-1])
					s++
					break
				}
			}
			// check snake eat self
			for i := 1; i < len(body); i++ {
				if body[0].X == body[i].X && body[0].Y == body[i].Y {
					gameOver = true
				}
			}
			// go ahead
			body = body[0 : len(body)-1]
			body = append([]*drawlib.Vector{body[0].Add(v)}, body...)
			// return snake when it out
			if body[0].X < -size {
				body[0].X = float64(int(w/size))*speed - speed
			} else if body[0].X > w {
				body[0].X = -speed
			} else if body[0].Y < -size {
				body[0].Y = float64(int(h/size))*speed - speed
			} else if body[0].Y > h {
				body[0].Y = -speed
			}
		} else {
			dc.DrawString("Game Over!", float64(d.Width()/2)-75, float64(d.Height()/2))
		}

	})

	d.OnKeyPress(func(k key.Code) {
		// change direction
		if k == key.CodeLeftArrow && v.X == 0 {
			v.X, v.Y = -speed, 0
		} else if k == key.CodeRightArrow && v.X == 0 {
			v.X, v.Y = speed, 0
		} else if k == key.CodeUpArrow && v.Y == 0 {
			v.X, v.Y = 0, -speed
		} else if k == key.CodeDownArrow && v.Y == 0 {
			v.X, v.Y = 0, speed
		}
	})
	d.Start()
}

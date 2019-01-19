package main

import (
	"github.com/ATTHDEV/drawlib"
)

func main() {

	d := drawlib.New(drawlib.Option().Title("Hello").Dimension(600, 600))
	w, h := float64(d.Width()), float64(d.Height())
	d.Render(func() {
		d.Canvas.Background(255)
		d.Canvas.FillRGBA255(255, 0, 0, 16)
		for i := 0; i < 360; i += 15 {
			d.Canvas.Push()
			d.Canvas.RotateAbout(drawlib.ToRadians(float64(i)), w/2, h/2)
			d.Canvas.DrawEllipse(w/2, h/2, w*7/16, h/8)
			d.Canvas.Fill()
			d.Canvas.Pop()
		}
	})
	// d.OnKeyPress(func(k key.Code) {
	// 	if k == key.CodeA {
	// 		d.SetFullScreen(false)
	// 	} else if k == key.CodeS {
	// 		d.SetFullScreen(true)
	// 	} else if k == key.CodeD {
	// 		d.SetMaximize(false)
	// 	} else if k == key.CodeF {
	// 		d.SetMaximize(true)
	// 	} else if k == key.CodeLeftArrow {
	// 		d.SetLocation(800, 0)
	// 	} else if k == key.CodeRightArrow {
	// 		d.SetSize(200, 200)
	// 	}
	// })
	// d.OnMousePress(func(b mouse.Button, x, y int) {
	// 	fmt.Println("press : ", b, x, y)
	// })
	// d.OnMouseIsPress(func(b mouse.Button, x, y int) {
	// 	fmt.Println("is press : ", b, x, y)
	// })
	// d.OnMouseRelease(func(b mouse.Button, x, y int) {
	// 	fmt.Println("release : ", b, x, y)
	// })
	// d.OnMouseWheel(func(step int, x, y int) {
	// 	fmt.Println(step, x, y)
	// })
	// d.OnMouseMove(func(x, y int) {
	// 	fmt.Println(x, y)
	// })
	d.Start()
	d.CaptureScreen("out")
}

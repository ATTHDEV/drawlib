package drawlib

import (
	"image"
	"image/color"
	"image/draw"
	"log"
	"sync"
	"time"

	"github.com/ATTHDEV/shiny/driver"
	"github.com/ATTHDEV/shiny/screen"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
)

var (
	tickDuration             = time.Second / 60
	defaultWindowsBackground = color.RGBA{240, 240, 240, 255}
)

type (
	updateEvent struct {
		dt float64
	}
	Drawlib struct {
		mutex                 *sync.Mutex
		options               *screen.WindowOptions
		buffer                screen.Buffer
		screen                screen.Screen
		window                screen.Window
		texture               screen.Texture
		rect                  image.Rectangle
		drawState             int8
		Canvas                *Canvas
		keyIsPress            bool
		keyIsPressCode        key.Code
		mouseIsPress          bool
		mouseIsPressButton    mouse.Button
		mouseIsPressX         int
		mouseIsPressY         int
		defaultCloseOperation bool
		autoscale             bool
		publish               bool
		renderCallback        *func()
		renderLoopCallback    *func(float64)
		sizeCallback          *func(int, int)
		KeyPressCallback      *func(key.Code)
		KeyReleaseCallback    *func(key.Code)
		keyIsPressCallback    *func(key.Code)
		mousePressCallback    *func(mouse.Button, int, int)
		mouseIsPressCallback  *func(mouse.Button, int, int)
		mouseReleaseCallback  *func(mouse.Button, int, int)
		mouseWheelCallback    *func(int, int, int)
		mouseMoveCallback     *func(int, int)
		visibleCallback       *func()
		hiddenCallback        *func()
		closeCallback         *func()
		initCallback          *func()
	}
)

func (d *Drawlib) SetAutoScale(value bool) {
	d.autoscale = value
}

func (d *Drawlib) SetDefualteCloseOperation(value bool) {
	d.defaultCloseOperation = value
}

func (d *Drawlib) Init(f func()) {
	d.initCallback = &f
}

func (d *Drawlib) Render(f func()) {
	d.renderCallback = &f
}

func (d *Drawlib) RenderLoop(f func(float64)) {
	d.renderLoopCallback = &f
}

func (d *Drawlib) OnSizeChange(f func(int, int)) {
	d.sizeCallback = &f
}

func (d *Drawlib) OnKeyPress(f func(key.Code)) {
	d.KeyPressCallback = &f
}

func (d *Drawlib) OnKeyRelease(f func(key.Code)) {
	d.KeyReleaseCallback = &f
}

func (d *Drawlib) OnKeyIsPress(f func(key.Code)) {
	d.keyIsPressCallback = &f
}

func (d *Drawlib) OnMousePress(f func(mouse.Button, int, int)) {
	d.mousePressCallback = &f
}

func (d *Drawlib) OnMouseRelease(f func(mouse.Button, int, int)) {
	d.mouseReleaseCallback = &f
}

func (d *Drawlib) OnMouseIsPress(f func(mouse.Button, int, int)) {
	d.mouseIsPressCallback = &f
}

func (d *Drawlib) OnMouseWheel(f func(int, int, int)) {
	d.mouseWheelCallback = &f
}

func (d *Drawlib) OnMouseMove(f func(int, int)) {
	d.mouseMoveCallback = &f
}

func (d *Drawlib) OnWindowsVisible(f func()) {
	d.visibleCallback = &f
}

func (d *Drawlib) OnWindowsHidden(f func()) {
	d.hiddenCallback = &f
}

func (d *Drawlib) OnWindowsClose(f func()) {
	d.closeCallback = &f
}

func New(o ...*screen.WindowOptions) *Drawlib {
	var options *screen.WindowOptions
	if len(o) == 1 {
		options = o[0]
	} else {
		options = screen.NewWindowOptions(
			screen.Title("Drawlib windows"),
			screen.Dimensions(600, 600),
		)
	}
	return &Drawlib{
		options:               options,
		Canvas:                NewCanvas(options.Width, options.Height),
		defaultCloseOperation: true,
	}
}

func (d *Drawlib) Start() {
	driver.Main(func(s screen.Screen) {
		w, err := s.NewWindow(d.options)

		if err != nil {
			log.Fatal(err)
		}
		defer w.Release()
		d.mutex = &sync.Mutex{}
		d.screen = s
		d.window = w
		d.rect = image.Rect(0, 0, d.options.Width, d.options.Height)

		d.buffer, err = s.NewBuffer(image.Point{d.options.Width, d.options.Height})
		if err != nil {
			panic(err)
		}

		d.texture, err = d.screen.NewTexture(d.buffer.Bounds().Max)
		if err != nil {
			panic(err)
		}

		if d.initCallback != nil {
			(*d.initCallback)()
		}

		go func() {
			ticker := time.NewTicker(tickDuration)
			timeStart := time.Now().UnixNano()
			var tickerC <-chan time.Time
			for {
				tickerC = ticker.C
				select {
				case <-tickerC:
					if d.keyIsPressCallback != nil {
						if d.keyIsPress {
							(*d.keyIsPressCallback)(d.keyIsPressCode)
						}
					}
					if d.mouseIsPressCallback != nil {
						if d.mouseIsPress {
							(*d.mouseIsPressCallback)(d.mouseIsPressButton, d.mouseIsPressX, d.mouseIsPressY)
						}
					}
					now := time.Now().UnixNano()
					delta := float64(now-timeStart) / 1000000000
					timeStart = now
					if d.renderLoopCallback != nil {
						(*d.renderLoopCallback)(delta)
					}
					w.Send(updateEvent{})
				}
			}
		}()
		d.eventLoop()
	})
}

func (d *Drawlib) eventLoop() {
	if d.renderCallback != nil {
		(*d.renderCallback)()
	}
	for {
		e := d.window.NextEvent()
		switch e := e.(type) {
		case lifecycle.Event:
			switch e.To {
			case lifecycle.StageDead:
				if d.closeCallback != nil {
					(*d.closeCallback)()
				}
				return
			case lifecycle.StageFocused:
				if d.visibleCallback != nil {
					(*d.visibleCallback)()
				}
			case lifecycle.StageVisible:
				if d.hiddenCallback != nil {
					(*d.hiddenCallback)()
				}
			}
		case key.Event:
			if d.defaultCloseOperation {
				if e.Code == key.CodeEscape {
					return
				}
			}
			switch e.Direction {
			case key.DirPress:
				d.keyIsPress = true
				d.keyIsPressCode = e.Code
				if d.KeyPressCallback != nil {
					(*d.KeyPressCallback)(e.Code)
				}
			case key.DirRelease:
				d.keyIsPress = false
				if d.KeyReleaseCallback != nil {
					(*d.KeyReleaseCallback)(e.Code)
				}
			}
		case mouse.Event:
			switch e.Direction {
			case mouse.DirPress:
				d.mouseIsPress = true
				d.mouseIsPressButton = e.Button
				if d.mousePressCallback != nil {
					(*d.mousePressCallback)(e.Button, int(e.X), int(e.Y))
				}
			case mouse.DirRelease:
				d.mouseIsPress = false
				d.mouseIsPressX = int(e.X)
				d.mouseIsPressY = int(e.Y)
				if d.mouseReleaseCallback != nil {
					(*d.mouseReleaseCallback)(e.Button, d.mouseIsPressX, d.mouseIsPressY)
				}
			case mouse.DirStep:
				if d.mouseWheelCallback != nil {
					if e.Button == -1 {
						(*d.mouseWheelCallback)(1, int(e.X), int(e.Y))
					} else if e.Button == -2 {
						(*d.mouseWheelCallback)(-1, int(e.X), int(e.Y))
					}
				}
			case mouse.DirNone:
				if d.mouseMoveCallback != nil {
					(*d.mouseMoveCallback)(int(e.X), int(e.Y))
				}
			}
		case paint.Event:
			// if d.renderCallback != nil {
			// 	(*d.renderCallback)()
			// }
		case size.Event:
			d.mutex.Lock()
			size := e.Size()
			d.options.Width = size.X
			d.options.Height = size.Y
			//fmt.Println(d.config.Width, d.config.Height)
			if d.autoscale {
				d.rect = e.Bounds()
			} else {
				// update canvas position
				w := d.Canvas.Width()
				h := d.Canvas.Height()
				if size.X >= w && size.Y >= h {
					offsetX := (size.X - d.Canvas.Width()) / 2
					offsetY := (size.Y - d.Canvas.Height()) / 2
					offsetW := offsetX + d.Canvas.Width()
					offsetH := offsetY + d.Canvas.Height()
					d.window.Fill(image.Rect(0, 0, offsetX, size.Y), defaultWindowsBackground, draw.Src)
					d.window.Fill(image.Rect(offsetW, 0, size.X, size.Y), defaultWindowsBackground, draw.Src)
					d.window.Fill(image.Rect(0, 0, size.X, offsetY), defaultWindowsBackground, draw.Src)
					d.window.Fill(image.Rect(0, offsetH, size.X, size.Y), defaultWindowsBackground, draw.Src)
					d.rect = image.Rect(offsetX, offsetY, offsetW, offsetH)

				} else if size.X < w || size.Y < h {
					if size.X < size.Y {
						offsetY := (size.Y-h)/2 + (w-size.X)/2
						offsetX := offsetY + size.X
						d.window.Fill(image.Rect(0, 0, size.X, offsetY), defaultWindowsBackground, draw.Src)
						d.window.Fill(image.Rect(0, offsetX, size.X, size.Y), defaultWindowsBackground, draw.Src)
						d.rect = image.Rect(0, offsetY, size.X, offsetX)

					} else {
						offsetX := (size.X-w)/2 + (h-size.Y)/2
						offsetY := offsetX + size.Y
						d.window.Fill(image.Rect(0, 0, offsetX, size.Y), defaultWindowsBackground, draw.Src)
						d.window.Fill(image.Rect(offsetY, 0, size.X, size.Y), defaultWindowsBackground, draw.Src)
						d.rect = image.Rect(offsetX, 0, offsetY, size.Y)
					}
				}
			}
			if d.sizeCallback != nil {
				(*d.sizeCallback)(size.X, size.Y)
			}
			d.mutex.Unlock()
		case updateEvent:
			d.swapbuffer()
		case error:
			log.Print(e)
		}
	}
}

func (d *Drawlib) swapbuffer() {
	d.mutex.Lock()
	draw.Draw(d.buffer.RGBA(), d.buffer.Bounds(), d.Canvas.im, image.ZP, draw.Src)
	d.texture.Upload(image.ZP, d.buffer, d.buffer.Bounds())
	d.window.Scale(d.rect, d.texture, d.texture.Bounds(), draw.Src, nil)
	d.window.Publish()
	d.mutex.Unlock()
}

func (d *Drawlib) CaptureScreen(path string) {
	d.Canvas.SavePNG(path + ".png")
}

func (d *Drawlib) Width() int {
	return d.options.Width
}

func (d *Drawlib) Height() int {
	return d.options.Height
}

func (d *Drawlib) Quit() {
	d.window.Send(lifecycle.Event{To: lifecycle.StageDead})
}

func (d *Drawlib) SetMaximize(maximize bool) {
	d.window.SetMaximize(maximize)
}

func (d *Drawlib) SetFullScreen(fullscreen bool) {
	d.window.SetFullScreen(fullscreen)
}

func (d *Drawlib) SetSize(width, height int) {
	d.window.SetDimention(int32(width), int32(height))
}

func (d *Drawlib) SetLocation(x, y int) {
	d.window.MoveWindow(int32(x), int32(y), int32(d.Width()), int32(d.Height()))
}

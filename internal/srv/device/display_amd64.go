package device

import (
	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"image"
	_ "image/png"
	"log"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/devices/v3/ssd1306"
	"sync"
)

type Display struct {
	oledLock    sync.Mutex
	oledDisplay *ssd1306.Dev
	i2cBus      i2c.BusCloser

	lock           sync.RWMutex
	on             bool
	simulationMode bool
	lastImg        image.Image

	simulationWindow *app.Window

	askDone chan bool
	askImg  chan image.Image
	done    chan bool
}

func (d *Display) startSimulation() {
	d.simulationWindow = app.NewWindow(app.Size(unit.Px(256), unit.Px(128)), app.MinSize(unit.Px(128), unit.Px(64)))
	go func() {
		if err := d.gioloop(); err != nil {
			log.Fatal(err)
		}
	}()
	go app.Main()
}

func (d *Display) invalidateSimulationWindow() {
	d.simulationWindow.Invalidate()
}

func (d *Display) closeSimulationWindow() {
	d.simulationWindow.Close()
}

func (d *Display) gioloop() error {
	var ops op.Ops
	for {
		e := <-d.simulationWindow.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)

			d.lock.RLock()
			lastImg := d.lastImg
			d.lock.RUnlock()

			img := widget.Image{Src: paint.NewImageOp(lastImg), Fit: widget.Contain}
			img.Layout(gtx)
			e.Frame(gtx.Ops)
		}
	}
}

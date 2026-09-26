package device

import (
	"image"
	_ "image/png"
	"log"
	"log/slog"
	"sync"

	"github.com/jypelle/vekigi/internal/tool"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/devices/v3/ssd1306"
	"periph.io/x/host/v3"
)

type Display struct {
	oledLock    sync.Mutex
	oledDisplay *ssd1306.Dev
	i2cBus      i2c.BusCloser

	lock           sync.RWMutex
	on             bool
	simulationMode bool
	lastImg        image.Image

	askDone chan bool
	askImg  chan image.Image
	done    chan bool
}

func NewDisplay(simulationMode bool) *Display {
	if !simulationMode {
		if _, err := host.Init(); err != nil {
			log.Fatal(err)
		}
	}

	device := Display{
		simulationMode: simulationMode,
		askDone:        make(chan bool),
		askImg:         make(chan image.Image),
		done:           make(chan bool),
	}

	return &device
}

func (d *Display) Start() {
	slog.Info("Start display device")

	d.on = true

	// In simulation mode, the simulator window reads the current image by itself
	if !d.simulationMode {
		var err error
		// Open a handle to the first available I²C bus:
		d.i2cBus, err = i2creg.Open("")
		if err != nil {
			tool.Fatal("Unable to open i2c bus", "error", err)
		}

		// Open a handle to a ssd1306 connected on the I²C bus:
		d.oledDisplay, err = ssd1306.NewI2C(d.i2cBus, &ssd1306.DefaultOpts)
		if err != nil {
			tool.Fatal("Unable to initialize oled display", "error", err)
		}

		d.oledDisplay.SetContrast(1)

		go func() {
			for loop := true; loop; {
				select {
				case <-d.askDone:
					loop = false
				case newImg := <-d.askImg:
					d.oledLock.Lock()
					d.oledDisplay.Draw(d.oledDisplay.Bounds(), newImg, image.Point{})
					d.oledLock.Unlock()
				}
			}
			d.oledLock.Lock()
			d.i2cBus.Close()
			d.oledLock.Unlock()
			d.done <- true
		}()
	}
}

func (d *Display) Stop() {
	slog.Info("Stop display device")

	if !d.simulationMode {
		d.askDone <- true
		<-d.done
	}

}

func (d *Display) SetOff() {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.setOff()
}

func (d *Display) setOff() {
	d.on = false
	if !d.simulationMode {
		d.oledLock.Lock()
		d.oledDisplay.Halt()
		d.oledLock.Unlock()
	}
}

func (d *Display) SetOn() {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.setOn()
}

func (d *Display) setOn() {
	d.on = true
	if !d.simulationMode {
		d.oledLock.Lock()
		d.oledDisplay.SetContrast(1) // Hack to force display on (calling Draw() is not enough)
		d.oledLock.Unlock()
		d.askImg <- d.lastImg
	}

}

func (d *Display) Switch() bool {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.on {
		d.setOff()
	} else {
		d.setOn()
	}

	return d.on
}

func (d *Display) IsOn() bool {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.on
}

func (d *Display) ShowImage(img image.Image) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.lastImg = img
	if d.on && !d.simulationMode {
		d.askImg <- img
	}
}

// image returns the image currently visible on the display, nil when the display is off
func (d *Display) image() image.Image {
	d.lock.RLock()
	defer d.lock.RUnlock()
	if !d.on {
		return nil
	}
	return d.lastImg
}

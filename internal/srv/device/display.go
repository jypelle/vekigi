package device

import (
	"github.com/sirupsen/logrus"
	"image"
	_ "image/png"
	"log"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/devices/v3/ssd1306"
	"periph.io/x/host/v3"
)

func NewDisplay(simulationMode bool) *Display {
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
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
	logrus.Infof("Start display device")

	d.on = true

	if d.simulationMode {
		d.startSimulation()
	} else {
		var err error
		// Open a handle to the first available I²C bus:
		d.i2cBus, err = i2creg.Open("")
		if err != nil {
			logrus.Fatalf("Unable to open i2c bus: %v\n", err)
		}

		// Open a handle to a ssd1306 connected on the I²C bus:
		d.oledDisplay, err = ssd1306.NewI2C(d.i2cBus, &ssd1306.DefaultOpts)
		if err != nil {
			logrus.Fatalf("Unable to initialize oled display: %v\n", err)
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
	logrus.Infof("Stop display device")

	if d.simulationMode {
		d.closeSimulationWindow()
	} else {
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
	if d.simulationMode {
		d.invalidateSimulationWindow()
	} else {
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
	if d.on {
		if d.simulationMode {
			d.invalidateSimulationWindow()
		} else {
			d.askImg <- img
		}
	}
}

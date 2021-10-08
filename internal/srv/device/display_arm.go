package device

import (
	"image"
	_ "image/png"
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

	askDone chan bool
	askImg  chan image.Image
	done    chan bool
}

func (d *Display) startSimulation() {
}

func (d *Display) invalidateSimulationWindow() {
}

func (d *Display) closeSimulationWindow() {
}

package device

import (
	"github.com/jypelle/vekigi/internal/srv/event"
	"github.com/sirupsen/logrus"
	"log"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"
	"sync"
	"time"
)

type Button struct {
	buttonId       event.ButtonId
	pin            gpio.PinIO
	isPressed      bool
	pressStepCount int64
	lastChange     time.Time
}

func NewButton(buttonId event.ButtonId, name string) *Button {
	button := Button{buttonId: buttonId, pin: gpioreg.ByName(name)}

	if button.pin == nil {
		logrus.Fatalf("Failed to find %s button", name)
	}

	// Set it as input, with an internal pull up resistor:
	if err := button.pin.In(gpio.PullUp, gpio.NoEdge); err != nil {
		logrus.Fatalf("Failed to setup %s button: %v", name, err)
	}
	return &button
}

func (b *Button) Refresh(buttonEventChannel chan event.ButtonEvent) {
	wasPressed := b.isPressed
	b.isPressed = bool(!b.pin.Read())

	now := time.Now()
	if !b.isPressed && wasPressed {
		b.lastChange = now
		buttonEventChannel <- event.ButtonEvent{ButtonId: b.buttonId, ButtonEventType: event.RELEASE_EVENT_TYPE, PressStepCount: b.pressStepCount}
		b.pressStepCount = 0
	} else if b.isPressed && b.lastChange.Add(160*time.Millisecond).Before(now) {
		b.lastChange = now
		b.pressStepCount++
		buttonEventChannel <- event.ButtonEvent{ButtonId: b.buttonId, ButtonEventType: event.PRESS_EVENT_TYPE, PressStepCount: b.pressStepCount}
	}
}

type Buttons struct {
	lock         sync.RWMutex
	eventChannel chan event.ButtonEvent
	simulation   bool

	buttons []*Button

	checkTicker *time.Ticker

	askDone chan bool
	done    chan bool
}

func NewButtons(simulation bool) *Buttons {
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	device := Buttons{
		eventChannel: make(chan event.ButtonEvent),
		simulation:   simulation,
		askDone:      make(chan bool),
		done:         make(chan bool),
	}

	return &device
}

func (d *Buttons) Start() {
	logrus.Infof("Start buttons device")

	d.lock.Lock()
	defer d.lock.Unlock()

	if !d.simulation {
		d.buttons = append(d.buttons, NewButton(event.DIGIT1_BUTTON, "GPIO16"))
		d.buttons = append(d.buttons, NewButton(event.DIGIT2_BUTTON, "GPIO13"))
		d.buttons = append(d.buttons, NewButton(event.DIGIT3_BUTTON, "GPIO12"))
		d.buttons = append(d.buttons, NewButton(event.DIGIT4_BUTTON, "GPIO6"))
		d.buttons = append(d.buttons, NewButton(event.DIGIT5_BUTTON, "GPIO5"))
		d.buttons = append(d.buttons, NewButton(event.DIGIT6_BUTTON, "GPIO23"))
		d.buttons = append(d.buttons, NewButton(event.PLAYLIST_BUTTON, "GPIO22"))
		d.buttons = append(d.buttons, NewButton(event.ALARM_SETTING_BUTTON, "GPIO27"))
		d.buttons = append(d.buttons, NewButton(event.LESS_BUTTON, "GPIO17"))
		d.buttons = append(d.buttons, NewButton(event.MORE_BUTTON, "GPIO4"))
		d.buttons = append(d.buttons, NewButton(event.SNOOZE_BUTTON, "GPIO25"))
		d.buttons = append(d.buttons, NewButton(event.NEXT_POWEROFF_BUTTON, "GPIO24"))
	}

	// Start periodic check
	d.checkTicker = time.NewTicker(5 * time.Millisecond)
	go func() {
		for loop := true; loop; {
			select {
			case <-d.checkTicker.C:
				for _, button := range d.buttons {
					button.Refresh(d.eventChannel)
				}
			case <-d.askDone:
				loop = false
			}
		}
		d.done <- true
	}()
}

func (d *Buttons) StopSendingEvent() {
	logrus.Infof("Stop buttons device")

	d.lock.Lock()
	defer d.lock.Unlock()

	d.checkTicker.Stop()
	d.askDone <- true
	<-d.done
	//close(d.eventChannel)
}

func (d *Buttons) EventChannel() chan event.ButtonEvent {
	return d.eventChannel
}

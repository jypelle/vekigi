package device

import (
	"github.com/jypelle/vekigi/internal/srv/config"
	"github.com/jypelle/vekigi/internal/srv/event"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

type Clock struct {
	lock         sync.RWMutex
	eventChannel chan event.TickerEvent

	serverConfig       *config.ServerConfig
	refreshClockTicker *time.Ticker

	snoozeWakeUpTimer *time.Timer

	askDone chan bool
	done    chan bool
}

func NewClock(serverConfig *config.ServerConfig) *Clock {
	ticker := Clock{
		eventChannel: make(chan event.TickerEvent),
		serverConfig: serverConfig,
		askDone:      make(chan bool),
		done:         make(chan bool),
	}
	return &ticker
}

func (d *Clock) Start() {
	logrus.Infof("Start ticker device")
	d.lock.Lock()
	defer d.lock.Unlock()

	d.refreshClockTicker = time.NewTicker(time.Second)

	go func() {
		var oldDisplayedTime string
		var oldTimerTickEventTime time.Time

		for loop := true; loop; {
			select {
			case <-d.refreshClockTicker.C:
				now := time.Now()

				// Check starting minute
				displayedTime := now.Format("15:04")
				if oldDisplayedTime != displayedTime {
					d.eventChannel <- event.TickerEvent{Data: event.TickerEventTickData{}}
				}
				oldDisplayedTime = displayedTime

				// Check Alarm
				alarmTime := d.serverConfig.Alarm()
				if alarmTime.Enabled &&
					int64(now.Hour()) == alarmTime.Hour &&
					int64(now.Minute()) == alarmTime.Minute &&
					(int64(oldTimerTickEventTime.Hour()) != alarmTime.Hour || int64(oldTimerTickEventTime.Minute()) != alarmTime.Minute) {

					d.TriggerAlarm()
				}
				oldTimerTickEventTime = now

			case <-d.askDone:
				loop = false
			}
		}
		d.done <- true
	}()
}

func (d *Clock) StopSendingEvent() {
	logrus.Infof("Stop ticker device")
	d.lock.Lock()
	defer d.lock.Unlock()

	d.refreshClockTicker.Stop()
	d.clearAlarm()
	d.askDone <- true
	<-d.done
	//close(d.eventChannel)
}

func (d *Clock) EventChannel() chan event.TickerEvent {
	return d.eventChannel
}

func (d *Clock) ClearAlarm() {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.clearAlarm()
}

func (d *Clock) clearAlarm() {
	if d.snoozeWakeUpTimer != nil {
		d.snoozeWakeUpTimer.Stop()
		d.snoozeWakeUpTimer = nil
	}
}

func (d *Clock) TriggerAlarm() {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.clearAlarm()
	d.snoozeWakeUpTimer = time.AfterFunc(0, func() {
		d.eventChannel <- event.TickerEvent{Data: event.TickerEventAlarmData{}}
	})
}

func (d *Clock) Snooze() {
	d.lock.Lock()
	defer d.lock.Unlock()
	if d.snoozeWakeUpTimer != nil {
		logrus.Infof("Snooze for %d seconds", d.serverConfig.SnoozeDuration)
		d.snoozeWakeUpTimer.Reset(time.Duration(d.serverConfig.SnoozeDuration) * time.Second)
	}
}

func (d *Clock) IsAlarmRunning() bool {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.snoozeWakeUpTimer != nil
}

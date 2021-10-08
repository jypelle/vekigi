package device

import (
	"fmt"
	"github.com/jypelle/vekigi/apimodel"
	"github.com/jypelle/vekigi/internal/srv/config"
	"github.com/jypelle/vekigi/internal/srv/event"
	"github.com/sirupsen/logrus"
	"os/exec"
	"sync"
)

type WebradioPlayer struct {
	lock         sync.RWMutex
	eventChannel chan event.WebradioEvent

	webradioGroups map[int64][]*config.Webradio

	currentRadioId  *apimodel.WebradioId
	currentRadioCmd *exec.Cmd

	sendEvent bool
}

func NewWebradioPlayer(config *config.ServerConfig) *WebradioPlayer {
	webradioPlayer := WebradioPlayer{
		webradioGroups: config.WebradioGroups,
		eventChannel:   make(chan event.WebradioEvent),
		sendEvent:      true,
	}
	return &webradioPlayer
}

func (d *WebradioPlayer) Start() {
	logrus.Infof("Start webradio player device")
}

func (d *WebradioPlayer) StopSendingEvent() {
	logrus.Infof("Stop sending events for webradio player device")

	d.lock.Lock()
	defer d.lock.Unlock()

	d.sendEvent = false
	//close(d.eventChannel)
}

func (d *WebradioPlayer) Stop() {
	logrus.Infof("Stop webradio player device")

	d.lock.Lock()
	defer d.lock.Unlock()

	d.clear()
}

func (d *WebradioPlayer) EventChannel() chan event.WebradioEvent {
	return d.eventChannel
}

func (d *WebradioPlayer) Play(radioId apimodel.WebradioId) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.currentRadioId != nil && radioId == *d.currentRadioId {
		logrus.Infof("Already listening radio %d", radioId)
		return nil
	}

	radioGroup, ok := d.webradioGroups[radioId.GroupId]
	if !ok {
		return fmt.Errorf("Radio group %d is undefined", radioId.GroupId)
	}

	if radioId.IndexId > int64(len(radioGroup)) {
		return fmt.Errorf("Radio %d for group %d is undefined", radioId.IndexId, radioId.GroupId)
	}
	webradio := radioGroup[radioId.IndexId-1]
	if webradio.Url == "" {
		return fmt.Errorf("Radio %d is undefined", radioId)
	}
	d.clear()

	logrus.Infof("Listening Radio %d: \"%s\" ", radioId, webradio.Name)
	d.currentRadioCmd = exec.Command("cvlc", "--aout=alsa", "--play-and-exit", webradio.Url)
	err := d.currentRadioCmd.Start()
	if err != nil {
		return fmt.Errorf("Unable to listen Radio %d", radioId)
	}
	d.currentRadioId = &radioId

	currentRadioCmd := d.currentRadioCmd
	go func() {
		_ = currentRadioCmd.Wait()
		d.lock.Lock()
		defer d.lock.Unlock()
		if d.currentRadioCmd == currentRadioCmd {
			d.currentRadioCmd = nil
			d.currentRadioId = nil
			if d.sendEvent {
				go func() { d.eventChannel <- event.WebradioEvent{Data: event.WebradioEventStopPlayingData{}} }()
			}
		}
	}()

	return nil
}

func (d *WebradioPlayer) CurrentWebRadio() *config.Webradio {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.currentRadioId == nil {
		return nil
	}

	return d.webradioGroups[d.currentRadioId.GroupId][d.currentRadioId.IndexId-1]
}

func (d *WebradioPlayer) Webradio(webradioId apimodel.WebradioId) *config.Webradio {
	d.lock.Lock()
	defer d.lock.Unlock()

	return d.webradio(webradioId)
}

func (d *WebradioPlayer) webradio(webradioId apimodel.WebradioId) *config.Webradio {
	radioGroup, ok := d.webradioGroups[webradioId.GroupId]
	if !ok {
		return nil
	}

	if webradioId.IndexId > int64(len(radioGroup)) {
		return nil
	}
	webradio := radioGroup[webradioId.IndexId-1]
	if webradio.Url == "" {
		return nil
	}

	return webradio
}

func (d *WebradioPlayer) Clear() {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.clear()
}

func (d *WebradioPlayer) clear() {
	if d.currentRadioCmd != nil {
		if err := d.currentRadioCmd.Process.Kill(); err != nil {
			logrus.Errorf("Failed to kill process: %v", err)
		}
		d.currentRadioCmd = nil
		d.currentRadioId = nil
	}
}

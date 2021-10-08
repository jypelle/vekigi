package config

import (
	"github.com/jypelle/vekigi/apimodel"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"sync"
	"time"
)

type ServerState struct {
	serverStateConfig     ServerStateConfig
	lock                  sync.RWMutex
	backupTimer           *time.Timer
	completeStateFilename string
}

func NewsServerState(completeStateFilename string) *ServerState {
	serverState := &ServerState{
		completeStateFilename: completeStateFilename,
	}

	rawConfig, err := ioutil.ReadFile(completeStateFilename)
	if err == nil {
		// Interpret state file
		err = yaml.Unmarshal(rawConfig, &serverState.serverStateConfig)
		if err != nil {
			logrus.Fatalf("Unable to interpret config file: %v\n", err)
		}
	} else {
		// Create default state file
		logrus.Infof("Create default state file")
		serverState.SetVolume(40)
		serverState.SetAlarm(Alarm{Hour: 8, Minute: 0})
	}

	return serverState
}

func (ss *ServerState) Volume() int64 {
	ss.lock.RLock()
	defer ss.lock.RUnlock()

	return ss.serverStateConfig.Volume
}

func (ss *ServerState) SetVolume(volume int64) {
	ss.lock.Lock()
	defer ss.lock.Unlock()

	ss.serverStateConfig.Volume = volume
	ss.scheduleSave()
}

func (ss *ServerState) Alarm() Alarm {
	ss.lock.RLock()
	defer ss.lock.RUnlock()

	return ss.serverStateConfig.Alarm
}

func (ss *ServerState) SetAlarm(alarm Alarm) {
	ss.lock.Lock()
	defer ss.lock.Unlock()

	ss.serverStateConfig.Alarm = alarm
	ss.scheduleSave()
}

func (ss *ServerState) scheduleSave() {
	if ss.backupTimer == nil {
		ss.backupTimer = time.AfterFunc(10*time.Second, func() {
			ss.lock.Lock()
			defer ss.lock.Unlock()
			ss.save()
		})
	} else {
		ss.backupTimer.Reset(10 * time.Second)
	}
}

func (ss *ServerState) save() {
	logrus.Infof("Save state file: %s", ss.completeStateFilename)
	rawConfig, err := yaml.Marshal(&ss.serverStateConfig)
	if err != nil {
		logrus.Fatalf("Unable to serialize state file: %v\n", err)
	}
	err = ioutil.WriteFile(ss.completeStateFilename, rawConfig, 0660)
	if err != nil {
		logrus.Fatalf("Unable to save state file: %v\n", err)
	}
}

func (ss *ServerState) FlushSave() {
	ss.lock.Lock()
	defer ss.lock.Unlock()
	if ss.backupTimer != nil {
		if ss.backupTimer.Stop() {
			ss.save()
		}
	}
}

type ServerStateConfig struct {
	Volume int64 `yaml:"volume"`
	Alarm  Alarm `yaml:"alarm"`
}

type Alarm struct {
	Hour       int64                `yaml:"hour"`
	Minute     int64                `yaml:"minute"`
	WebradioId *apimodel.WebradioId `yaml:"webradio_id"`
	PlaylistId *apimodel.PlaylistId `yaml:"playlist_id"`
	Enabled    bool                 `yaml:"enabled"`
}

func (sc *Alarm) AddMinute(minutes int64) {
	sc.Minute += minutes
	if sc.Minute >= 60 {
		sc.Hour += sc.Minute / 60
		sc.Minute = sc.Minute % 60
	} else if sc.Minute < 0 {
		sc.Hour += sc.Minute/60 - 1
		sc.Minute = sc.Minute%60 + 60
	}
	if sc.Hour >= 24 {
		sc.Hour = sc.Hour % 24
	} else if sc.Hour < 0 {
		sc.Hour = sc.Hour%24 + 24
	}
	logrus.Debugf("New alarm value: %02d:%02d", sc.Hour, sc.Minute)
}

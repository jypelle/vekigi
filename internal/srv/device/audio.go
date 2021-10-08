package device

import (
	"github.com/jypelle/vekigi/internal/srv/config"
	"github.com/sirupsen/logrus"
	"os/exec"
	"strconv"
	"sync"
)

type Audio struct {
	lock         sync.RWMutex
	serverState  *config.ServerState
	zeroSoundCmd *exec.Cmd
}

func NewAudio(serverState *config.ServerState) *Audio {
	device := Audio{serverState: serverState}
	return &device
}

func (w *Audio) Start() {
	logrus.Infof("Start audio device")

	w.lock.Lock()
	defer w.lock.Unlock()

	w.zeroSoundCmd = exec.Command("aplay", "-D", "default", "-t", "raw", "-r", "44100", "-c", "2", "-f", "S16_LE", "/dev/zero")
	err := w.zeroSoundCmd.Start()
	if err != nil {
		logrus.Panic("Unable to activate popping/clicking cleaner: %v", err)
		return
	}

	w.applyVolume()
}

func (w *Audio) Stop() {
	logrus.Infof("Stop audio device")

	w.lock.Lock()
	defer w.lock.Unlock()

	if err := w.zeroSoundCmd.Process.Kill(); err != nil {
		logrus.Errorf("Failed to stop popping/clicking cleaner: %v", err)
	}
}

func (w *Audio) setVolume(volume int64) {
	if volume > 100 {
		volume = 100
	}
	if volume < 0 {
		volume = 0
	}
	w.serverState.SetVolume(volume)
	w.applyVolume()
}

func (w *Audio) applyVolume() {
	cmd := exec.Command("amixer", "set", "PCM", strconv.FormatInt(int64(w.serverState.Volume()), 10)+"%")
	err := cmd.Run()
	if err != nil {
		logrus.Warnf("Unable to set volume")
		return
	}
}

func (w *Audio) IncreaseVolume() {
	logrus.Infof("Increase volume")
	w.lock.Lock()
	defer w.lock.Unlock()
	w.setVolume(w.serverState.Volume() + 4)
}

func (w *Audio) DecreaseVolume() {
	logrus.Infof("Decrease volume")
	w.lock.Lock()
	defer w.lock.Unlock()
	w.setVolume(w.serverState.Volume() - 4)
}

func (w *Audio) SetVolume(volume int64) error {
	logrus.Infof("Set volume")
	w.lock.Lock()
	defer w.lock.Unlock()
	w.setVolume(volume)
	return nil
}

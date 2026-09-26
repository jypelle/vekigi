package device

import (
	"fmt"
	"log/slog"
	"os/exec"
	"strconv"
	"sync"

	"github.com/jypelle/vekigi/internal/srv/config"
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
	slog.Info("Start audio device")

	w.lock.Lock()
	defer w.lock.Unlock()

	w.zeroSoundCmd = exec.Command("aplay", "-D", "default", "-t", "raw", "-r", "44100", "-c", "2", "-f", "S16_LE", "/dev/zero")
	err := w.zeroSoundCmd.Start()
	if err != nil {
		panic(fmt.Errorf("unable to activate popping/clicking cleaner: %w", err))
	}

	w.applyVolume()
}

func (w *Audio) Stop() {
	slog.Info("Stop audio device")

	w.lock.Lock()
	defer w.lock.Unlock()

	if err := w.zeroSoundCmd.Process.Kill(); err != nil {
		slog.Error("Failed to stop popping/clicking cleaner", "error", err)
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
		slog.Warn("Unable to set volume", "error", err)
		return
	}
}

func (w *Audio) IncreaseVolume() {
	slog.Info("Increase volume")
	w.lock.Lock()
	defer w.lock.Unlock()
	w.setVolume(w.serverState.Volume() + 4)
}

func (w *Audio) DecreaseVolume() {
	slog.Info("Decrease volume")
	w.lock.Lock()
	defer w.lock.Unlock()
	w.setVolume(w.serverState.Volume() - 4)
}

func (w *Audio) SetVolume(volume int64) error {
	slog.Info("Set volume")
	w.lock.Lock()
	defer w.lock.Unlock()
	w.setVolume(volume)
	return nil
}

package srv

import (
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"os/exec"
	"time"

	"github.com/gorilla/mux"
	"github.com/jypelle/vekigi/internal/srv/config"
	"github.com/jypelle/vekigi/internal/srv/device"
	"github.com/jypelle/vekigi/internal/srv/event"
	"github.com/jypelle/vekigi/internal/version"
)

type ServerApp struct {
	*config.ServerConfig
	displayDevice        *device.Display
	audioDevice          *device.Audio
	webradioPlayerDevice *device.WebradioPlayer
	playlistPlayerDevice device.PlaylistPlayer
	clockDevice          *device.Clock
	buttonsDevice        *device.Buttons
	apiDevice            *device.Api

	currentMode Mode

	currentPopUp   PopUp
	popUpHideTimer *time.Timer

	animationTickCount int
	animationTickTimer *time.Timer

	internalEventChannel chan event.InternalEvent

	eventLoopAskDone chan bool
	eventLoopDone    chan bool

	apiRouter *mux.Router
}

type Mode int64

const (
	UNDEFINED_MODE Mode = iota
	CLOCK_MODE
	ALARM_SETTING_MODE
	END_MODE
)

type PopUp int64

const (
	NO_POPUP PopUp = iota
	VOLUME_POPUP
	SNOOZE_OFF_POPUP
)

func NewServerApp(configDir string, debugMode bool) *ServerApp {

	slog.Debug("Creation of vekigi server ...", "version", version.AppVersion.String())

	app := &ServerApp{
		currentMode:          UNDEFINED_MODE,
		internalEventChannel: make(chan event.InternalEvent),
		eventLoopAskDone:     make(chan bool),
		eventLoopDone:        make(chan bool),
		ServerConfig:         config.NewServerConfig(configDir, debugMode),
	}

	app.displayDevice = device.NewDisplay(app.SimulationMode)
	app.audioDevice = device.NewAudio(app.ServerState)
	app.webradioPlayerDevice = device.NewWebradioPlayer(app.ServerConfig)
	if app.ServerConfig.MifasolParam == nil {
		app.playlistPlayerDevice = device.NewLocalPlaylistPlayer(app.ServerConfig.GetCompletePlaylistFolder())
	} else {
		app.playlistPlayerDevice = device.NewMifasolPlaylistPlayer(app.ServerConfig.MifasolParam)
	}
	app.clockDevice = device.NewClock(app.ServerConfig)
	app.buttonsDevice = device.NewButtons(app.SimulationMode)
	app.apiDevice = device.NewApi(app.ServerConfig)

	slog.Debug("Server created")

	return app
}

func (s *ServerApp) Start() {
	slog.Info("Starting vekigi server ...")

	// Init random generator
	rand.Seed(time.Now().UnixNano())

	slog.Info("Starting devices ...")

	// Start display device
	s.displayDevice.Start()

	// Display startup screen
	s.refreshDisplay(true)
	time.Sleep(2 * time.Second)

	// Start volume device
	s.audioDevice.Start()

	// Start webradio player device
	s.webradioPlayerDevice.Start()

	// Start playlist player device
	s.playlistPlayerDevice.Start()

	// Start event loop
	go s.eventLoop()

	// Start clock device
	s.clockDevice.Start()

	// Start buttons device
	s.buttonsDevice.Start()

	// Start api device
	s.apiDevice.Start()

	// Set clock mode
	s.currentMode = CLOCK_MODE
	s.refreshDisplay(true)

}

// WaitForStop blocks until a stop signal is received. When built with -tags simulator, the simulator window runs
// meanwhile on the calling goroutine (which must be the main one) and closing it also returns (with a nil signal)
func (s *ServerApp) WaitForStop(signals <-chan os.Signal) os.Signal {
	return device.WaitForStop(s.displayDevice, signals)
}

func (s *ServerApp) Stop(halt bool) {
	slog.Info("Stopping vekigi server ...")

	// Stop api
	s.apiDevice.StopSendingEvent()

	// Stop buttons device
	s.buttonsDevice.StopSendingEvent()

	// Stop clock device
	s.clockDevice.StopSendingEvent()

	// Stop playlist player event
	s.playlistPlayerDevice.StopSendingEvent()

	// Stop webradio player event
	s.webradioPlayerDevice.StopSendingEvent()

	// Stop event loop
	slog.Info("Stop event loop")
	s.eventLoopAskDone <- true
	<-s.eventLoopDone

	// Display end mode image
	s.currentMode = END_MODE
	s.refreshDisplay(true)

	// Stop playlist player
	s.playlistPlayerDevice.Stop()

	// Stop webradio player
	s.webradioPlayerDevice.Stop()

	// Stop volume device
	s.audioDevice.Stop()

	// Stop display device
	s.displayDevice.Stop()

	// Flush config backup
	s.ServerConfig.ServerState.FlushSave()

	slog.Info("Server stopped")

	if halt {
		slog.Info("System halt")
		haltCmd := exec.Command("sudo", "halt")
		err := haltCmd.Run()
		if err != nil {
			panic(fmt.Errorf("unable to halt the system: %w", err))
		}
	}
	os.Exit(0)
}

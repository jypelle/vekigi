//go:build simulator

package device

import (
	"image"
	"log/slog"
	"os"
	"sync/atomic"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jypelle/vekigi/internal/srv/event"
	"github.com/jypelle/vekigi/internal/tool"
)

const simulatorScale = 4

// Keyboard keys simulating each button. Keys are physical positions, chosen to be the same on QWERTY and AZERTY layouts
var simulatorKeys = map[event.ButtonId][]ebiten.Key{
	event.DIGIT1_BUTTON:        {ebiten.Key1, ebiten.KeyNumpad1},
	event.DIGIT2_BUTTON:        {ebiten.Key2, ebiten.KeyNumpad2},
	event.DIGIT3_BUTTON:        {ebiten.Key3, ebiten.KeyNumpad3},
	event.DIGIT4_BUTTON:        {ebiten.Key4, ebiten.KeyNumpad4},
	event.DIGIT5_BUTTON:        {ebiten.Key5, ebiten.KeyNumpad5},
	event.DIGIT6_BUTTON:        {ebiten.Key6, ebiten.KeyNumpad6},
	event.PLAYLIST_BUTTON:      {ebiten.KeyP},
	event.ALARM_SETTING_BUTTON: {ebiten.KeyEnter, ebiten.KeyNumpadEnter},
	event.LESS_BUTTON:          {ebiten.KeyArrowDown, ebiten.KeyNumpadSubtract},
	event.MORE_BUTTON:          {ebiten.KeyArrowUp, ebiten.KeyNumpadAdd},
	event.SNOOZE_BUTTON:        {ebiten.KeySpace},
	event.NEXT_POWEROFF_BUTTON: {ebiten.KeyArrowRight},
}

// Pressed state of each simulated button, written by the window loop and read by the buttons device
var simulatorPressed [event.NEXT_POWEROFF_BUTTON + 1]atomic.Bool

func simulatorButtons() []*Button {
	var buttons []*Button
	for buttonId := range simulatorPressed {
		buttons = append(buttons, &Button{
			buttonId:    event.ButtonId(buttonId),
			readPressed: simulatorPressed[buttonId].Load,
		})
	}
	return buttons
}

type simulatorGame struct {
	display *Display
	signals <-chan os.Signal
	signal  os.Signal

	srcImg image.Image
	img    *ebiten.Image
}

func (g *simulatorGame) Update() error {
	select {
	case g.signal = <-g.signals:
		return ebiten.Termination
	default:
	}

	for buttonId, keys := range simulatorKeys {
		pressed := false
		for _, key := range keys {
			pressed = pressed || ebiten.IsKeyPressed(key)
		}
		simulatorPressed[buttonId].Store(pressed)
	}
	return nil
}

func (g *simulatorGame) Draw(screen *ebiten.Image) {
	srcImg := g.display.image()
	if srcImg == nil {
		return
	}
	if srcImg != g.srcImg {
		if g.img != nil {
			g.img.Deallocate()
		}
		g.srcImg = srcImg
		g.img = ebiten.NewImageFromImage(srcImg)
	}
	screen.DrawImage(g.img, nil)
}

func (g *simulatorGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 128, 64
}

// WaitForStop runs the simulator window until a stop signal is received (returned) or the window is closed (nil returned).
// It must be called from the main goroutine
func WaitForStop(display *Display, signals <-chan os.Signal) os.Signal {
	slog.Info("Simulator keys: 1-6 digits, P playlist, Enter alarm setting, Up/Down more/less, Space snooze, Right next/power off")

	ebiten.SetWindowTitle("Vekigi simulator")
	ebiten.SetWindowSize(128*simulatorScale, 64*simulatorScale)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := &simulatorGame{display: display, signals: signals}
	if err := ebiten.RunGame(game); err != nil {
		tool.Fatal("Simulator failure", "error", err)
	}
	if game.signal == nil {
		slog.Info("Simulator window closed")
	}
	return game.signal
}

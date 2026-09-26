//go:build !simulator

package device

import "os"

func simulatorButtons() []*Button {
	return nil
}

// WaitForStop blocks until a stop signal is received
func WaitForStop(display *Display, signals <-chan os.Signal) os.Signal {
	return <-signals
}

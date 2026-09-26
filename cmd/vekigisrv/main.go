package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/jypelle/vekigi/internal/srv"
	"github.com/jypelle/vekigi/internal/version"
	"github.com/lmittmann/tint"
)

const configSuffix = "vekigi"

func main() {

	// Logger
	slog.SetDefault(newLogger(slog.LevelInfo, "2006-01-02T15:04:05"))

	mainCommand := filepath.Base(os.Args[0])

	// region Flags and Commands definition

	// Debug Mode
	debugMode := flag.Bool("d", false, "Enable debug mode")

	// User config dir
	defaultConfigDir := "./." + configSuffix
	userConfigDir, err := os.UserConfigDir()
	os.UserCacheDir()
	if err == nil {
		defaultConfigDir = filepath.Join(userConfigDir, configSuffix)
	}
	configDir := flag.String("c", defaultConfigDir, "Location of vekigi config folder")

	// Usage
	flag.Usage = func() {
		fmt.Printf("\nUsage: %s [OPTIONS] [COMMAND]\n", mainCommand)
		fmt.Printf("\nA webradio alarm clock\n")
		fmt.Printf("\nOptions:\n")
		flag.PrintDefaults()
		fmt.Printf("\nCommands:\n")
		fmt.Printf("  run       Run server\n")
		fmt.Printf("  version   Show the version number\n")
		fmt.Printf("\nRun '%s COMMAND --help' for more information on a command.\n", mainCommand)
	}

	// run command
	runCmd := flag.NewFlagSet("run", flag.ExitOnError)

	runCmd.Usage = func() {
		fmt.Printf("\nUsage: %s run\n", mainCommand)
		fmt.Printf("\nRun the server\n")
	}

	// version command
	versionCmd := flag.NewFlagSet("version", flag.ExitOnError)

	versionCmd.Usage = func() {
		fmt.Printf("\nUsage: %s version\n", mainCommand)
		fmt.Printf("\nShow the version information\n")
	}

	// endregion

	// region Flags and Commands Parsing
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(0)
	}

	switch flag.Arg(0) {
	case "run":
		runCmd.Parse(flag.Args()[1:])
		if runCmd.NArg() > 0 {
			fmt.Printf("\n\"%s %s\" accepts no arguments\n", mainCommand, flag.Arg(0))
			runCmd.Usage()
			os.Exit(1)
		}
	case "version":
		versionCmd.Parse(flag.Args()[1:])
		if versionCmd.NArg() > 0 {
			fmt.Printf("\n\"%s %s\" accepts no arguments\n", mainCommand, flag.Arg(0))
			versionCmd.Usage()
			os.Exit(1)
		}
	default:
		fmt.Printf("\n%s is not a vekigi command\n", flag.Args()[0])
		flag.Usage()
		os.Exit(1)
	}
	// endregion

	if *debugMode {
		slog.SetDefault(newLogger(slog.LevelDebug, time.RFC3339Nano))
		slog.Info("Debug mode activated")
	}

	// Create vekigi server
	serverApp := srv.NewServerApp(*configDir, *debugMode)

	if versionCmd.Parsed() {
		fmt.Printf("Version %s\n", version.AppVersion.String())
	} else {
		if runCmd.Parsed() {
			// Listen stop signal
			ch := make(chan os.Signal, 1)
			signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGABRT, syscall.SIGHUP, syscall.SIGUSR1)

			// Start vekigi server
			serverApp.Start()

			sig := serverApp.WaitForStop(ch)
			if sig != nil {
				slog.Info("Received signal", "signal", sig)
			}
			serverApp.Stop(sig == syscall.SIGUSR1)
		}
	}

}

func newLogger(level slog.Level, timeFormat string) *slog.Logger {
	return slog.New(
		tint.NewTextHandler(os.Stdout, &tint.Options{
			Level:      level,
			TimeFormat: timeFormat,
		}),
	)
}

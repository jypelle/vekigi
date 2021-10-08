package main

import (
	"flag"
	"fmt"
	"github.com/jypelle/vekigi/internal/srv"
	"github.com/jypelle/vekigi/internal/version"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

const configSuffix = "vekigi"

func main() {

	// Logger
	logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})

	mainCommand := filepath.Base(os.Args[0])

	// region Flags and Commands definition

	// Debug Mode
	debugMode := flag.Bool("d", false, "Enable debug mode")

	// Simulation Mode
	simulationMode := flag.Bool("s", false, "Enable simulation mode")

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
		logrus.SetLevel(logrus.DebugLevel)
		logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true, FullTimestamp: true, TimestampFormat: time.RFC3339Nano})
		logrus.Printf("Debug mode activated")
	}

	// Create vekigi server
	serverApp := srv.NewServerApp(*configDir, *debugMode, *simulationMode)

	if versionCmd.Parsed() {
		fmt.Printf("Version %s\n", version.AppVersion.String())
	} else {
		if runCmd.Parsed() {
			// Listen stop signal
			ch := make(chan os.Signal)
			signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGABRT, syscall.SIGHUP, syscall.SIGUSR1)

			// Start pensees server
			serverApp.Start()

			sig := <-ch
			logrus.Infof("Received signal: %v", sig)
			serverApp.Stop(sig == syscall.SIGUSR1)
		}
	}

}

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"github.com/jacobsa/go-serial/serial"

	"github.com/omaskery/rn2483"
)

var CLI struct {
	Port      string `kong:"arg,help='serial device port to open'"`
	Verbosity int    `kong:"short='v',type='counter',help='increases the logging verbosity',env='VERBOSITY'"`
	BaudRate  uint   `kong:"default='57600',help='baud rate for serial port',env='BAUDRATE'"`

	AssumeText bool `kong:"help='assumes packets are printable text, and prints them as such',env='ASSUME_TEXT'"`
}

func main() {
	kong.Parse(
		&CLI,
		kong.Description("simple test program to listen for incoming packets"),
		kong.UsageOnError(),
	)

	logger := stdr.New(log.Default())
	stdr.SetVerbosity(CLI.Verbosity)

	if err := errMain(logger); err != nil {
		logger.Error(err, "program exiting with error")
		os.Exit(1)
	}
}

func errMain(logger logr.Logger) error {
	s, err := serial.Open(serial.OpenOptions{
		PortName:        CLI.Port,
		BaudRate:        CLI.BaudRate,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 1,
	})
	if err != nil {
		return fmt.Errorf("error opening serial port: %v", err)
	}

	dbg := &rn2483.DebugSerial{
		Serial:     s,
		Logger:     logger.WithName("serial").V(1),
		AssumeText: true,
	}

	device := rn2483.New(rn2483.Config{
		Serial: dbg,
	})
	defer func() {
		if err := device.Close(); err != nil {
			logger.Error(err, "error closing device")
		}
	}()

	pauseDuration, err := device.PauseMAC()
	if err != nil {
		return fmt.Errorf("error pausing MAC layer: %w", err)
	}
	logger.Info("MAC layer paused", "duration", pauseDuration, "until", time.Now().Add(pauseDuration))

	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGKILL)

	go func() {
		for {
			rx, err := device.RadioRx(rn2483.ContinuousReceiveMode)
			if errors.Is(err, rn2483.ErrReceiveTimeout) {
				continue
			}
			if err != nil {
				logger.Error(err, "error receiving packet")
				break
			}

			data := string(rx)
			if !CLI.AssumeText {
				data = rn2483.BytesToHex(rx)
			}
			logger.Info("received", "data", data)
		}
	}()

	<-stop

	return nil
}

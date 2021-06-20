package main

import (
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
	"go.uber.org/multierr"

	"github.com/omaskery/rn2483"
)

var CLI struct {
	Port      string `kong:"arg,help='serial device port to open'"`
	Verbosity int    `kong:"short='v',type='counter',help='increases the logging verbosity',env='VERBOSITY'"`
	BaudRate  uint   `kong:"default='57600',help='baud rate for serial port',env='BAUDRATE'"`
}

func main() {
	kong.Parse(
		&CLI,
		kong.Description("simple example that blinks the user LEDs on and off"),
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

	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGKILL)

	defer func() {
		if errs := multierr.Combine(
			device.SetDigitalGPIO(rn2483.PinGPIO10, false),
			device.SetDigitalGPIO(rn2483.PinGPIO11, false),
		); errs != nil {
			logger.Error(err, "error turning off user LEDs")
		}
	}()

	ledState := true
	for {
		if err := device.SetDigitalGPIO(rn2483.PinGPIO10, ledState); err != nil {
			return fmt.Errorf("error toggling user LED (GPIO10): %v", err)
		}
		if err := device.SetDigitalGPIO(rn2483.PinGPIO11, !ledState); err != nil {
			return fmt.Errorf("error toggling user LED (GPIO11): %v", err)
		}
		ledState = !ledState

		select {
		case <-time.After(time.Second):
		case <-stop:
			return nil
		}
	}
}

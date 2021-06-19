package main

import (
	"fmt"
	"log"
	"os"

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
}

func main() {
	kong.Parse(
		&CLI,
		kong.Description("retrieve and display information about the connected device"),
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

	v, err := device.GetVersion()
	if err != nil {
		return fmt.Errorf("error getting device version: %w", err)
	}
	logger.Info("device version", "SKU", v.SKU, "version", v.VersionString(), "release-time", v.ReleaseTime, "is-known-sku", v.IsKnownSKU())

	vdd, err := device.GetVDD()
	if err != nil {
		return fmt.Errorf("error getting device VDD: %w", err)
	}
	logger.Info("device VDD", "millivolts", vdd.Millivolts(), "volts", vdd.Volts())

	hweui, err := device.GetHWEUI()
	if err != nil {
		return fmt.Errorf("error getting device HWEUI: %w", err)
	}
	logger.Info("device HWEUI", "hweui", hweui)

	for _, param := range rn2483.KnownRadioParameters {
		value, err := device.GetRadioParameter(param)
		if err != nil {
			return fmt.Errorf("error getting radio parameter %s: %w", param, err)
		}
		logger.Info("radio parameter", param, value)
	}

	return nil
}

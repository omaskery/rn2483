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

	"github.com/omaskery/rn2483"
)

var CLI struct {
	Port      string `kong:"arg,help='serial device port to open'"`
	Verbosity int    `kong:"short='v',type='counter',help='increases the logging verbosity',env='VERBOSITY'"`
	BaudRate  uint   `kong:"default='57600',help='baud rate for serial port',env='BAUDRATE'"`

	RadioPower int `kong:"default='10',help='set the radio power when transmitting',env='RADIO_POWER'"`

	TransmitInterval time.Duration `kong:"default='10s',help='interval between transmissions',env='TRANSMIT_INTERVAL'"`
}

func main() {
	kong.Parse(
		&CLI,
		kong.Description("simple test program to transmit a test packet at a fixed interval"),
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

	if err := device.SetRadioPower(CLI.RadioPower); err != nil {
		return fmt.Errorf("error setting radio power: %w", err)
	}

	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGKILL)

	counter := 0
	for {
		transmitTime := time.Now().UTC().Format(time.RFC3339Nano)
		packetNumber := counter
		counter++

		packet := fmt.Sprintf("test: pkt=%d ts=%s", packetNumber, transmitTime)

		logger.Info("transmitting", "packet", packet)
		if err := device.RadioTx([]byte(packet)); err != nil {
			logger.Error(err, "failed to transmit packet", "packet", packetNumber, "ts", transmitTime)
			continue
		}

		select {
		case <-time.After(CLI.TransmitInterval):
		case <-stop:
			return nil
		}
	}
}

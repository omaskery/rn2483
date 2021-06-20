package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"github.com/jacobsa/go-serial/serial"
)

var CLI struct {
	Port      string `kong:"arg,help='serial device port to open'"`
	Verbosity int    `kong:"short='v',type='counter',help='increases the logging verbosity',env='VERBOSITY'"`
	BaudRate  uint   `kong:"default='57600',help='baud rate for serial port',env='BAUDRATE'"`
}

func main() {
	kong.Parse(
		&CLI,
		kong.Description("primitive terminal interface for sending commands to the device"),
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
	logger.V(1).Info("opening serial port", "port", CLI.Port)
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
	defer func() {
		if err := s.Close(); err != nil {
			logger.Error(err, "error closing serial port")
		}
	}()

	lines := make(chan string, 10)
	go func() {
		output := bufio.NewScanner(s)
		for output.Scan() {
			lines <- strings.TrimSpace(output.Text())
		}
	}()

	input := bufio.NewScanner(os.Stdin)

	logger.V(1).Info("processing terminal input")
	fmt.Printf("> ")
	for input.Scan() {
		tx := input.Text()

		if tx == "quit" {
			break
		}

		switch tx {
		case "quit":
			break
		default:
			if _, err := fmt.Fprintf(s, "%s\r\n", strings.TrimSpace(tx)); err != nil {
				fmt.Printf("ERR: %v\n", err)
			}
		}

		deadline := time.After(1 * time.Second)
	outer:
		for {
			select {
			case line := <-lines:
				fmt.Printf("RX: %s\n", line)
			case <-deadline:
				break outer
			}
		}

		fmt.Printf("> ")
	}

	return nil
}

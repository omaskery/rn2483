package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/alecthomas/kong"
	"github.com/augustoroman/hexdump"
	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"github.com/jacobsa/go-serial/serial"

	"github.com/omaskery/rn2483"
)

var CLI struct {
	Port      string `kong:"flag,required,help='serial device port to open',env='PORT'"`
	Verbosity int    `kong:"short='v',type='counter',help='increases the logging verbosity',env='VERBOSITY'"`
	BaudRate  uint   `kong:"default='57600',help='baud rate for serial port',env='BAUDRATE'"`

	Read  cmdRead  `kong:"cmd,help='read data from NVM'"`
	Write cmdWrite `kong:"cmd,help='write data to NVM'"`
}

type cmdContext struct {
	Logger logr.Logger
	Device *rn2483.Device
}

func main() {
	cmd := kong.Parse(
		&CLI,
		kong.Description("allows for viewing and manipulating user NVM data"),
		kong.UsageOnError(),
	)

	logger := stdr.New(log.Default())
	stdr.SetVerbosity(CLI.Verbosity)

	s, err := serial.Open(serial.OpenOptions{
		PortName:        CLI.Port,
		BaudRate:        CLI.BaudRate,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 1,
	})
	if err != nil {
		logger.Error(err, "error opening serial port")
		os.Exit(1)
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

	ctx := cmdContext{
		Logger: logger,
		Device: device,
	}

	if err := cmd.Run(ctx); err != nil {
		logger.Error(err, "program exiting with error")
		os.Exit(1)
	}
}

type outputFormat string

const (
	outputFormatRaw     outputFormat = "raw"
	outputFormatHexDump outputFormat = "hexdump"
)

type cmdRead struct {
	From   uint16       `kong:"default='0',help='address to read from NVM, 0 means from start (0x300)',env='FROM'"`
	Length uint16       `kong:"default='0',help='number of bytes to read from NVM, 0 means read to end (0x3FF)',env='LENGTH'"`
	Format outputFormat `kong:"default='raw',choice='raw,hexdump',help='choose how the output is presented',env='FORMAT'"`
}

func (cmd *cmdRead) Run(ctx cmdContext) error {
	from := rn2483.UserNVMStart
	if cmd.From != 0 {
		from = cmd.From
	}

	length := (rn2483.UserNVMEnd + 1) - from
	if cmd.Length != 0 {
		length = cmd.Length
	}

	ctx.Logger.Info("reading from NVM", "start-address", rn2483.UInt16ToHex(from), "length", rn2483.UInt16ToHex(length))
	data, err := rn2483.ReadNVM(ctx.Device, from, length)
	if err != nil {
		return err
	}

	switch cmd.Format {
	case outputFormatRaw:
		if _, err := os.Stdout.Write(data); err != nil {
			return fmt.Errorf("error writing NVM data to stdout: %w", err)
		}
	case outputFormatHexDump:
		fmt.Println(hexdump.Dump(data))
	}

	return nil
}

type cmdWrite struct {
	To   uint16 `kong:"default='0',help='address to write to in NVM',env='TO'"`
	Data string `kong:"type='existingfile',default='-',help='file containing data to write to NVM',env='DATA'"`
}

func (cmd *cmdWrite) Run(ctx cmdContext) error {
	var data []byte
	var err error

	switch cmd.Data {
	case "-":
		data, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("error reading data from stdin: %w", err)
		}
	default:
		data, err = ioutil.ReadFile(cmd.Data)
		if err != nil {
			return fmt.Errorf("error reading data from file: %w", err)
		}
	}

	writeStart := rn2483.UserNVMStart
	if cmd.To != 0 {
		writeStart = cmd.To
	}

	writeEnd := writeStart + uint16(len(data))
	if writeStart < rn2483.UserNVMStart || writeEnd > rn2483.UserNVMEnd {
		return fmt.Errorf(
			"write would be out of bounds (%s->%s exceeds %s->%s)",
			rn2483.UInt16ToHex(writeStart), rn2483.UInt16ToHex(writeEnd),
			rn2483.UInt16ToHex(rn2483.UserNVMStart), rn2483.UInt16ToHex(rn2483.UserNVMEnd),
		)
	}

	if err := rn2483.WriteNVM(ctx.Device, writeStart, data); err != nil {
		return err
	}

	return nil
}

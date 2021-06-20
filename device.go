package rn2483

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

var (
	ErrInvalidParam    = errors.New("invalid parameter")
	ErrUnknown         = errors.New("unknown error")
	ErrTransceiverBusy = errors.New("the transceiver is currently busy")
)

// Config allows for configuring a new Device
type Config struct {
	Serial io.ReadWriteCloser
}

// Device represents a single RN2483 (or 2903) device, providing methods for configuring and querying its state and
// invoking the various features of the device (primarily transmitting & receiving packets)
type Device struct {
	serial io.ReadWriteCloser
	reader *bufio.Reader
}

// New creates a new Device
func New(cfg Config) *Device {
	return &Device{
		serial: cfg.Serial,
		reader: bufio.NewReader(cfg.Serial),
	}
}

// Close shuts the underlying serial device
func (d *Device) Close() error {
	if err := d.serial.Close(); err != nil {
		return fmt.Errorf("error closing serial device: %w", err)
	}

	return nil
}

func encodeBoolean(b bool) int {
	if b {
		return 1
	}

	return 0
}

// Sendf allows for easily sending arbitrary commands to the device
func (d *Device) Sendf(format string, a ...interface{}) error {
	command := fmt.Sprintf(format, a...)
	command = command + "\r\n"

	_, err := d.serial.Write([]byte(command))
	if err != nil {
		return fmt.Errorf("error writing to serial device: %w", err)
	}

	return nil
}

// ReadResponse allows for easily reading a line of text from the device in response to a command
func (d *Device) ReadResponse() (string, error) {
	line, err := d.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error reading from serial device: %w", err)
	}

	line = strings.TrimSpace(line)

	return line, nil
}

// ExecuteCommand sends the provided command, then reads and returns the response
func (d *Device) ExecuteCommand(format string, a ...interface{}) (string, error) {
	if err := d.Sendf(format, a...); err != nil {
		return "", err
	}

	line, err := d.ReadResponse()
	if err != nil {
		return "", err
	}

	return line, nil
}

// ExecuteCommandChecked sends the provided command, reads the response, checks it for common error codes, then
// returns the response
func (d *Device) ExecuteCommandChecked(format string, a ...interface{}) (string, error) {
	line, err := d.ExecuteCommand(format, a...)
	if err != nil {
		return "", err
	}

	if err := CheckCommandResponse(line, true); err != nil {
		return "", err
	}

	return line, nil
}

// ExecuteCommandCheckedStrict sends the provided command, reads the response and then checks it is one of a set of
// known command responses, failing any unrecognised responses
func (d *Device) ExecuteCommandCheckedStrict(format string, a ...interface{}) error {
	line, err := d.ExecuteCommand(format, a...)
	if err != nil {
		return err
	}

	if err := CheckCommandResponse(line, false); err != nil {
		return err
	}

	return nil
}

// CheckCommandResponse validates a command response against known error codes, optionally failing unknown responses
func CheckCommandResponse(line string, allowUnknown bool) error {
	mapping := map[string]error{
		"ok":            nil,
		"invalid_param": ErrInvalidParam,
		"busy":          ErrTransceiverBusy,
	}

	if err, ok := mapping[line]; ok {
		return err
	}

	if !allowUnknown {
		return fmt.Errorf("%w: %s", ErrUnknown, line)
	}

	return nil
}

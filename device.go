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

type Config struct {
	Serial io.ReadWriteCloser
}

type Device struct {
	serial io.ReadWriteCloser
	reader *bufio.Reader
}

func New(cfg Config) *Device {
	return &Device{
		serial: cfg.Serial,
		reader: bufio.NewReader(cfg.Serial),
	}
}

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

func (d *Device) Sendf(format string, a ...interface{}) error {
	command := fmt.Sprintf(format, a...)
	command = command + "\r\n"

	_, err := d.serial.Write([]byte(command))
	if err != nil {
		return fmt.Errorf("error writing to serial device: %w", err)
	}

	return nil
}

func (d *Device) ReadResponse() (string, error) {
	line, err := d.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error reading from serial device: %w", err)
	}

	line = strings.TrimSpace(line)

	return line, nil
}

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

func (d *Device) ExecuteCommandChecked(format string, a ...interface{}) error {
	line, err := d.ExecuteCommand(format, a...)
	if err != nil {
		return err
	}

	if err := CheckCommandResponse(line); err != nil {
		return err
	}

	return nil
}

func CheckCommandResponse(line string) error {
	switch line {
	case "ok":
	case "invalid_param":
		return ErrInvalidParam
	case "busy":
		return ErrTransceiverBusy
	default:
		return fmt.Errorf("%w: %s", ErrUnknown, line)
	}

	return nil
}

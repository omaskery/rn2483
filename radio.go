package rn2483

import (
	"errors"
	"fmt"
	"strings"
)

var KnownRadioParameters = []string{
	"bt", "mod", "freq", "pwr", "sf", "afcbw", "rxbw", "bitrate",
	"fdev", "prlen", "crc", "iqi", "cr", "wdt", "bw", "snr",
}

func (d *Device) SetRadioParameter(name string, value interface{}) error {
	return d.ExecuteCommandChecked("radio set %s %v", name, value)
}

func (d *Device) GetRadioParameter(name string) (string, error) {
	return d.ExecuteCommand("radio get %s", name)
}

func (d *Device) SetRadioPower(power int) error {
	return d.SetRadioParameter("pwr", power)
}

var (
	ErrTransmitTimeout = errors.New("transmission unsuccessful, interrupted by radio WDT")
)

func (d *Device) RadioTx(data []byte) error {
	err := d.ExecuteCommandChecked("radio tx %s", BytesToHex(data))
	if err != nil {
		return err
	}

	line, err := d.ReadResponse()
	if err != nil {
		return fmt.Errorf("error reading transmission result: %w", err)
	}

	switch line {
	case "radio_tx_ok":
	case "radio_err":
		return ErrTransmitTimeout
	default:
		return fmt.Errorf("%w: %s", ErrUnknown, line)
	}

	return nil
}

var (
	ErrReceiveTimeout = errors.New("reception unsuccessful, timeout occurred")

	ContinuousReceiveMode uint16 = 0
)

func (d *Device) RadioRx(windowSize uint16) ([]byte, error) {
	err := d.ExecuteCommandChecked("radio rx %d", windowSize)
	if err != nil {
		return nil, err
	}

	line, err := d.ReadResponse()
	if err != nil {
		return nil, fmt.Errorf("error reading receive result: %w", err)
	}

	if line == "radio_err" {
		return nil, ErrReceiveTimeout
	}

	const radioRxPrefix = "radio_rx"
	if !strings.HasPrefix(line, radioRxPrefix) {
		return nil, fmt.Errorf("%w: %s", ErrUnknown, line)
	}

	data := strings.TrimSpace(line[len(radioRxPrefix):])

	return HexToBytes(PadHexToEvenLength(data))
}

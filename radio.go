package rn2483

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// KnownRadioParameters lists the known parameters that can be get/set with radio commands
var KnownRadioParameters = []string{
	"bt", "mod", "freq", "pwr", "sf", "afcbw", "rxbw", "bitrate",
	"fdev", "prlen", "crc", "iqi", "cr", "wdt", "bw", "snr",
}

// SetRadioParameter sets the specified radio parameter to the given value
func (d *Device) SetRadioParameter(name string, value interface{}) error {
	return d.ExecuteCommandCheckedStrict("radio set %s %v", name, value)
}

// GetRadioParameter gets the specified radio parameter as a raw string value (direct from the response message)
func (d *Device) GetRadioParameter(name string) (string, error) {
	return d.ExecuteCommandChecked("radio get %s", name)
}

// SetRadioPower sets the radio's transmit power
func (d *Device) SetRadioPower(power int) error {
	return d.SetRadioParameter("pwr", power)
}

// GetRadioPower gets the radio's current configured transmit power
func (d *Device) GetRadioPower() (int, error) {
	valueStr, err := d.GetRadioParameter("pwr")
	if err != nil {
		return 0, err
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("error parsing radio power: %w", err)
	}

	return value, nil
}

var (
	ErrTransmitTimeout = errors.New("transmission unsuccessful, interrupted by radio WDT")
)

// RadioTx attempts to transmit a packet of data using the radio's current configuration
func (d *Device) RadioTx(data []byte) error {
	err := d.ExecuteCommandCheckedStrict("radio tx %s", BytesToHex(data))
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

// RadioRx causes the device to listen for, and return, a single packet. The window size determines how long it will
// listen for, behaving differently depending on the current radio configuration (particularly modulation mode).
func (d *Device) RadioRx(windowSize uint16) ([]byte, error) {
	err := d.ExecuteCommandCheckedStrict("radio rx %d", windowSize)
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

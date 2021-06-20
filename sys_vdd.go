package rn2483

import (
	"fmt"
)

// Voltage represents a voltage measured by the device
type Voltage uint16

// Volts is the measured voltage in Volts (V) as a decimal value
func (v Voltage) Volts() float64 {
	return float64(v) / 1000.0
}

// Millivolts is the measured voltage in Millivolts (mV) as an integer
func (v Voltage) Millivolts() uint16 {
	return uint16(v)
}

// GetVDD retrieves the current VDD voltage measured by the device
func (d *Device) GetVDD() (Voltage, error) {
	line, err := d.ExecuteCommandChecked("sys get vdd")
	if err != nil {
		return 0, err
	}

	var v Voltage
	if _, err := fmt.Sscanf(line, "%d", &v); err != nil {
		return 0, fmt.Errorf("error parsing voltage: %w", err)
	}

	return v, nil
}

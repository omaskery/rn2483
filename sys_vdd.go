package rn2483

import (
	"fmt"
)

type Voltage uint16

func (v Voltage) Volts() float64 {
	return float64(v) / 1000.0
}

func (v Voltage) Millivolts() uint16 {
	return uint16(v)
}

func (d *Device) GetVDD() (Voltage, error) {
	line, err := d.ExecuteCommand("sys get vdd")
	if err != nil {
		return 0, err
	}

	var v Voltage
	if _, err := fmt.Sscanf(line, "%d", &v); err != nil {
		return 0, fmt.Errorf("error parsing voltage: %w", err)
	}

	return v, nil
}

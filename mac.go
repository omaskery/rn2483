package rn2483

import (
	"fmt"
	"strconv"
	"time"
)

// PauseMAC pauses the LoRaWAN stack to allow transceiver configuration.
func (d *Device) PauseMAC() (time.Duration, error) {
	line, err := d.ExecuteCommandChecked("mac pause")
	if err != nil {
		return 0, err
	}

	milliseconds, err := strconv.Atoi(line)
	if err != nil {
		return 0, fmt.Errorf("error parsing pause duration: %w", err)
	}

	return time.Duration(milliseconds) * time.Millisecond, nil
}

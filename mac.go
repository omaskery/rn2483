package rn2483

import (
	"fmt"
	"strconv"
	"time"
)

func (d *Device) PauseMAC() (time.Duration, error) {
	line, err := d.ExecuteCommand("mac pause")
	if err != nil {
		return 0, err
	}

	milliseconds, err := strconv.Atoi(line)
	if err != nil {
		return 0, fmt.Errorf("error parsing pause duration: %w", err)
	}

	return time.Duration(milliseconds) * time.Millisecond, nil
}

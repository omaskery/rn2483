package rn2483

import (
	"fmt"
)

type GPIO int

const (
	GPIO10 GPIO = iota
	GPIO11
)

var GPIONames = map[GPIO]string{
	GPIO10: "GPIO10",
	GPIO11: "GPIO11",
}

func (d *Device) SetDigitalGPIO(gpio GPIO, value bool) error {
	gpioName := GPIONames[gpio]
	if gpioName == "" {
		panic(fmt.Sprintf("unknown GPIO: %v", gpio))
	}

	return d.ExecuteCommandChecked("sys set pindig %s %d", gpioName, encodeBoolean(value))
}


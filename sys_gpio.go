package rn2483

type PinName string

const (
	PinGPIO00  PinName = "GPIO00"
	PinGPIO01  PinName = "GPIO01"
	PinGPIO02  PinName = "GPIO02"
	PinGPIO03  PinName = "GPIO03"
	PinGPIO04  PinName = "GPIO04"
	PinGPIO05  PinName = "GPIO05"
	PinGPIO06  PinName = "GPIO06"
	PinGPIO07  PinName = "GPIO07"
	PinGPIO08  PinName = "GPIO08"
	PinGPIO09  PinName = "GPIO09"
	PinGPIO10  PinName = "GPIO10"
	PinGPIO11  PinName = "GPIO11"
	PinGPIO12  PinName = "GPIO12"
	PinGPIO13  PinName = "GPIO13"
	PinGPIO14  PinName = "GPIO14"
	PinUARTCTS PinName = "UART_CTS"
	PinUARTRTS PinName = "UART_RTS"
	PinTEST0   PinName = "TEST0"
	PinTEST1   PinName = "TEST1"
)

var AllPins = []PinName{
	PinGPIO00,
	PinGPIO01,
	PinGPIO02,
	PinGPIO03,
	PinGPIO04,
	PinGPIO05,
	PinGPIO06,
	PinGPIO07,
	PinGPIO08,
	PinGPIO09,
	PinGPIO10,
	PinGPIO11,
	PinGPIO12,
	PinGPIO13,
	PinGPIO14,
	PinUARTCTS,
	PinUARTRTS,
	PinTEST0,
	PinTEST1,
}

func (d *Device) SetDigitalGPIO(gpio PinName, value bool) error {
	return d.ExecuteCommandCheckedStrict("sys set pindig %s %d", string(gpio), encodeBoolean(value))
}

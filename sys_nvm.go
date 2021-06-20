package rn2483

import (
	"bytes"
	"fmt"
)

const (
	UserNVMStart uint16 = 0x300
	UserNVMEnd   uint16 = 0x3FF
)

var (
	UserNVMLength = (UserNVMEnd + 1) - UserNVMStart
)

func (d *Device) WriteNVM(address uint16, value byte) error {
	return d.ExecuteCommandCheckedStrict("sys set nvm %s %s", UInt16ToHex(address), ByteToHex(value))
}

func (d *Device) ReadNVM(address uint16) (byte, error) {
	line, err := d.ExecuteCommandChecked("sys get nvm %s", UInt16ToHex(address))
	if err != nil {
		return 0, err
	}

	return HexToByte(PadHexToEvenLength(line))
}

func ReadNVM(d *Device, start, amount uint16) ([]byte, error) {
	var buffer bytes.Buffer

	for i := start; i < (start + amount); i++ {
		value, err := d.ReadNVM(i)
		if err != nil {
			return nil, fmt.Errorf("error reading NVM at %s: %w", UInt16ToHex(i), err)
		}
		buffer.WriteByte(value)
	}

	return buffer.Bytes(), nil
}

func WriteNVM(d *Device, address uint16, data []byte) error {
	for i, b := range data {
		writeAddress := address + uint16(i)
		if err := d.WriteNVM(writeAddress, b); err != nil {
			return fmt.Errorf("error writing NVM at %s: %w", UInt16ToHex(writeAddress), err)
		}
	}

	return nil
}

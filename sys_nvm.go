package rn2483

import (
	"bytes"
	"fmt"
)

const (
	// UserNVMStart is the address of the first byte in user NVM that can be addressed
	UserNVMStart uint16 = 0x300
	// UserNVMEnd is the last address in user NVM that can be addressed
	UserNVMEnd uint16 = 0x3FF
)

var (
	// UserNVMLength is the number of bytes of user NVM that can be addressed
	UserNVMLength = (UserNVMEnd + 1) - UserNVMStart
)

// WriteNVM writes a single byte of data to user NVM
func (d *Device) WriteNVM(address uint16, value byte) error {
	return d.ExecuteCommandCheckedStrict("sys set nvm %s %s", UInt16ToHex(address), ByteToHex(value))
}

// ReadNVM reads a single byte of data from user NVM
func (d *Device) ReadNVM(address uint16) (byte, error) {
	line, err := d.ExecuteCommandChecked("sys get nvm %s", UInt16ToHex(address))
	if err != nil {
		return 0, err
	}

	return HexToByte(PadHexToEvenLength(line))
}

// ReadNVM reads a block of data from user NVM
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

// WriteNVM writes a block of data to user NVM
func WriteNVM(d *Device, address uint16, data []byte) error {
	for i, b := range data {
		writeAddress := address + uint16(i)
		if err := d.WriteNVM(writeAddress, b); err != nil {
			return fmt.Errorf("error writing NVM at %s: %w", UInt16ToHex(writeAddress), err)
		}
	}

	return nil
}

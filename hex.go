package rn2483

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
)

// BytesToHex returns the hex representation of a byte array
func BytesToHex(b []byte) string {
	return hex.EncodeToString(b)
}

// ByteToHex returns the hex representation of a single byte
func ByteToHex(b byte) string {
	return BytesToHex([]byte{b})
}

// UInt16ToHex returns the big-endian hex representation of a uint16
func UInt16ToHex(v uint16) string {
	var buffer bytes.Buffer
	_ = binary.Write(&buffer, binary.BigEndian, v)
	return BytesToHex(buffer.Bytes())
}

// HexToByte decodes a hex string into a single byte
func HexToByte(h string) (byte, error) {
	b, err := HexToBytes(h)
	if err != nil {
		return 0, err
	}
	if len(b) != 1 {
		return 0, errors.New("expected only one bytes worth of hex in input")
	}

	return b[0], nil
}

// HexToBytes decodes a hex string into a byte array
func HexToBytes(h string) ([]byte, error) {
	return hex.DecodeString(h)
}

// HexToUInt16 decodes a hex string of a big-endian uint16
func HexToUInt16(h string) (uint16, error) {
	b, err := HexToBytes(h)
	if err != nil {
		return 0, err
	}

	var result uint16
	if err := binary.Read(bytes.NewReader(b), binary.BigEndian, &result); err != nil {
		return 0, err
	}

	return result, nil
}

// PadHex pads a hexadecimal string with leading 0s until it satisfied the desired number of digits
func PadHex(h string, digits int) string {
	for len(h) < digits {
		h = "0" + h
	}
	return h
}

// PadHexToEvenLength pads a hex string with a single leading 0 only if the input hex string is an odd length
func PadHexToEvenLength(h string) string {
	if len(h) % 2 == 0 {
		return h
	}

	return "0" + h
}

package rn2483

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
)

func BytesToHex(b []byte) string {
	return hex.EncodeToString(b)
}

func ByteToHex(b byte) string {
	return BytesToHex([]byte{b})
}

func UInt16ToHex(v uint16) string {
	var buffer bytes.Buffer
	_ = binary.Write(&buffer, binary.BigEndian, v)
	return BytesToHex(buffer.Bytes())
}

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

func HexToBytes(h string) ([]byte, error) {
	return hex.DecodeString(h)
}

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

func PadHex(h string, digits int) string {
	for len(h) < digits {
		h = "0" + h
	}
	return h
}

func PadHexToEvenLength(h string) string {
	if len(h) % 2 == 0 {
		return h
	}

	return "0" + h
}

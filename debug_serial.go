package rn2483

import (
	"io"

	"github.com/go-logr/logr"
)

// DebugSerial wraps a serial device and logs data flowing to/from that device
type DebugSerial struct {
	Serial     io.ReadWriteCloser
	Logger     logr.Logger

	// AssumeText controls whether data flowing through this device is likely printable text, if true then
	// logged data will be printed as text, otherwise it is treated as unprintable binary and is displayed
	// however the logging library chooses to render []byte (often base64)
	AssumeText bool
}

// Read implements the io.ReadWriteCloser interface by calling the underlying Serial implementation and logging the read data
func (d *DebugSerial) Read(p []byte) (n int, err error) {
	n, err = d.Serial.Read(p)
	if n > 0 {
		d.Logger.Info("rx", "n", n, "data", d.prepareData(p[:n]))
	}
	return
}

// Write implements the io.ReadWriteCloser interface by calling the underlying Serial implementation and logging the sent data
func (d *DebugSerial) Write(p []byte) (n int, err error) {
	d.Logger.Info("tx", "data", d.prepareData(p))
	return d.Serial.Write(p)
}

// Close implements the io.ReadWriteCloser interface by calling the underlying Serial implementation
func (d *DebugSerial) Close() error {
	return d.Serial.Close()
}

var _ io.ReadWriteCloser = (*DebugSerial)(nil)

func (d *DebugSerial) prepareData(data []byte) interface{} {
	if d.AssumeText {
		return string(data)
	}

	return data
}

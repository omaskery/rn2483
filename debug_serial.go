package rn2483

import (
	"io"

	"github.com/go-logr/logr"
)

type DebugSerial struct {
	Serial     io.ReadWriteCloser
	Logger     logr.Logger
	AssumeText bool
}

func (d *DebugSerial) Read(p []byte) (n int, err error) {
	n, err = d.Serial.Read(p)
	if n > 0 {
		d.Logger.Info("rx", "n", n, "data", d.prepareData(p[:n]))
	}
	return
}

func (d *DebugSerial) Write(p []byte) (n int, err error) {
	d.Logger.Info("tx", "data", d.prepareData(p))
	return d.Serial.Write(p)
}

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

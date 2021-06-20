package fake

import (
	"fmt"
	"math/rand"

	"github.com/omaskery/rn2483"
)

type SysState struct {
	FirmwareVersion string
	GPIO            map[rn2483.PinName]bool
	NVM             []byte
}

func (s *SysState) ensureDefaults() {
	if s.FirmwareVersion == "" {
		s.FirmwareVersion = "RN2483 1.0.4 Mar 23 1991 13:37:00"
	}

	if s.GPIO == nil {
		s.GPIO = map[rn2483.PinName]bool{}
		for _, pin := range rn2483.AllPins {
			s.GPIO[pin] = false
		}
	}

	if s.NVM == nil {
		s.NVM = make([]byte, rn2483.UserNVMLength)
		for i := range s.NVM {
			s.NVM[i] = 0xFF
		}
	}
}

func (s *SysState) ReadNVM(address uint16) (byte, error) {
	if address < rn2483.UserNVMStart {
		return 0, fmt.Errorf("address out of bounds (%x before %x)", address, rn2483.UserNVMStart)
	}
	if address > rn2483.UserNVMEnd {
		return 0, fmt.Errorf("address out of bounds (%x after %x)", address, rn2483.UserNVMEnd)
	}

	index := address - rn2483.UserNVMStart
	return s.NVM[index], nil
}

func (s *SysState) WriteNVM(address uint16, value byte) error {
	if address < rn2483.UserNVMStart {
		return fmt.Errorf("address out of bounds (%x before %x)", address, rn2483.UserNVMStart)
	}
	if address > rn2483.UserNVMEnd {
		return fmt.Errorf("address out of bounds (%x after %x)", address, rn2483.UserNVMEnd)
	}

	index := address - rn2483.UserNVMStart
	s.NVM[index] = value

	return nil
}

func (d *Device) processSysCommand(ctx *commandContext, params []string) error {
	if len(params) < 1 {
		return invalidParam(ctx)
	}

	switch params[0] {
	case "get":
		return d.processSysGetCommand(ctx, params[1:])
	case "set":
		return d.processSysSetCommand(ctx, params[1:])
	case "reset":
		return ctx.WriteResponse(d.Sys.FirmwareVersion)
	default:
		return invalidParam(ctx)
	}
}

func (d *Device) processSysSetCommand(ctx *commandContext, params []string) error {
	if len(params) < 1 {
		return invalidParam(ctx)
	}

	switch params[0] {
	case "nvm":
		if len(params) < 3 {
			return invalidParam(ctx)
		}

		address, err := rn2483.HexToUInt16(rn2483.PadHex(params[1], 4))
		if err != nil {
			return invalidParam(ctx)
		}

		value, err := rn2483.HexToByte(params[2])
		if err != nil {
			return invalidParam(ctx)
		}

		if err := d.Sys.WriteNVM(address, value); err != nil {
			return invalidParam(ctx)
		}

		return ok(ctx)
	case "pindig":
		if len(params) < 3 {
			return invalidParam(ctx)
		}

		pinName := params[1]
		pinStateStr := params[2]

		if _, ok := d.Sys.GPIO[rn2483.PinName(pinName)]; !ok {
			return invalidParam(ctx)
		}

		if pinStateStr != "0" && pinStateStr != "1" {
			return invalidParam(ctx)
		}

		pinState := false
		if pinStateStr == "1" {
			pinState = true
		}

		d.Sys.GPIO[rn2483.PinName(pinName)] = pinState

		return ok(ctx)
	default:
		return invalidParam(ctx)
	}
}

func (d *Device) processSysGetCommand(ctx *commandContext, params []string) error {
	if len(params) < 1 {
		return invalidParam(ctx)
	}

	switch params[0] {
	case "ver":
		return ctx.WriteResponse(d.Sys.FirmwareVersion)
	case "vdd":
		voltage := 3304 + (rand.Intn(8) - 4)
		return ctx.WriteResponse("%d", voltage)
	case "nvm":
		if len(params) < 2 {
			return invalidParam(ctx)
		}

		address, err := rn2483.HexToUInt16(rn2483.PadHex(params[1], 4))
		if err != nil {
			return invalidParam(ctx)
		}

		value, err := d.Sys.ReadNVM(address)
		if err != nil {
			return invalidParam(ctx)
		}

		return ctx.WriteResponse(rn2483.ByteToHex(value))
	default:
		return invalidParam(ctx)
	}
}

package fake

import (
	"fmt"
	"strconv"
	"time"

	"github.com/omaskery/rn2483"
)

// RadioState holds the state of the fake device in relation to radio commands
type RadioState struct {
	// Power is the configured transmit power
	Power int

	// Tx is a callback invoked when the radio is asked to transmit a packet of data
	Tx func(d *Device, packet []byte) error
	// Rx is a callback invoked when the radio attempts to receive a packet of data, the function should
	// return a channel that may eventually yield a packet of data
	// TODO: should probably add a mechanism for the caller to release any resources associated with the callee
	Rx func(d *Device) <-chan []byte
}

func (r *RadioState) ensureDefaults() {
	r.Power = 10
}

func (d *Device) processRadioCommand(ctx *commandContext, params []string) error {
	if len(params) < 1 {
		return invalidParam(ctx)
	}

	switch params[0] {
	case "get":
		return d.processRadioGetCommand(ctx, params[1:])
	case "set":
		return d.processRadioSetCommand(ctx, params[1:])
	case "tx":
		return d.processRadioTxCommand(ctx, params[1:])
	case "rx":
		return d.processRadioRxCommand(ctx, params[1:])
	default:
		return invalidParam(ctx)
	}
}

func (d *Device) processRadioTxCommand(ctx *commandContext, params []string) error {
	if len(params) < 1 {
		return invalidParam(ctx)
	}

	data, err := rn2483.HexToBytes(params[0])
	if err != nil {
		return invalidParam(ctx)
	}

	if err := ok(ctx); err != nil {
		return fmt.Errorf("error sending initial transmit OK response: %w", err)
	}

	if d.Radio.Tx != nil {
		if err := d.Radio.Tx(d, data); err != nil {
			d.logger.Error(err, "radio transmit function returned an error")
			return ctx.writeResponse("radio_err")
		}
	} else {
		d.logger.Info("no transmit function registered: dropping transmission")
	}

	return ctx.writeResponse("radio_tx_ok")
}

func (d *Device) processRadioRxCommand(ctx *commandContext, params []string) error {
	if len(params) < 1 {
		return invalidParam(ctx)
	}

	rxWindow, err := strconv.Atoi(params[0])
	if err != nil {
		return invalidParam(ctx)
	}

	if err := ok(ctx); err != nil {
		return fmt.Errorf("error sending initial receive OK response: %w", err)
	}

	var rxChannel <-chan []byte

	if d.Radio.Rx != nil {
		rxChannel = d.Radio.Rx(d)
	} else {
		d.logger.Info("no receive function registered: will never receive data")
	}

	select {
	case <-time.After(time.Duration(rxWindow) * time.Millisecond):
		return ctx.writeResponse("radio_err")
	case data := <-rxChannel:
		return ctx.writeResponse("radio_rx %s", rn2483.BytesToHex(data))
	}
}

func (d *Device) processRadioGetCommand(ctx *commandContext, params []string) error {
	if len(params) < 1 {
		return invalidParam(ctx)
	}

	switch params[0] {
	case "pwr":
		return ctx.writeResponse("%d", d.Radio.Power)
	default:
		return invalidParam(ctx)
	}
}

func (d *Device) processRadioSetCommand(ctx *commandContext, params []string) error {
	if len(params) < 1 {
		return invalidParam(ctx)
	}

	switch params[0] {
	case "pwr":
		if len(params) < 2 {
			return invalidParam(ctx)
		}

		power, err := strconv.Atoi(params[1])
		if err != nil {
			return invalidParam(ctx)
		}

		d.Radio.Power = power

		return ok(ctx)
	default:
		return invalidParam(ctx)
	}
}

package fake

import (
	"time"
)

var (
	MaxPauseDuration = 4294967295 * time.Millisecond
)

type MacState struct {
	PausedUntil *time.Time
}

func (m *MacState) IsPaused() bool {
	return m.PausedUntil != nil && m.PausedUntil.After(time.Now())
}

func (m *MacState) Pause() time.Duration {
	pauseDuration := MaxPauseDuration
	pausedUntil := time.Now().Add(MaxPauseDuration)
	m.PausedUntil = &pausedUntil
	return pauseDuration
}

func (m *MacState) ensureDefaults() {
	m.PausedUntil = nil
}

func (d *Device) processMacCommand(ctx *commandContext, params []string) error {
	if len(params) < 1 {
		return invalidParam(ctx)
	}

	switch params[0] {
	case "pause":
		duration := d.Mac.Pause()
		return ctx.WriteResponse("%d", duration.Milliseconds())
	default:
		return invalidParam(ctx)
	}
}

package fake

import (
	"time"
)

var (
	// MaxPauseDuration is the maximum amount of time the mac layer can be paused
	MaxPauseDuration = 4294967295 * time.Millisecond
)

// MacState holds the state of the device in relation to mac commands
type MacState struct {
	// PausedUntil represents when, if paused, the MAC layer will un-pause
	PausedUntil *time.Time
}

// IsPaused determines whether the MAC layer is currently paused
func (m *MacState) IsPaused() bool {
	return m.PausedUntil != nil && m.PausedUntil.After(time.Now())
}

// Pause pauses the MAC layer, allowing for direct access to radio commands, returning the duration it will be paused for
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
		return ctx.writeResponse("%d", duration.Milliseconds())
	default:
		return invalidParam(ctx)
	}
}

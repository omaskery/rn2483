package ether

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/go-logr/logr"
	"github.com/jonboulle/clockwork"

	"github.com/omaskery/rn2483/fake"
)

var (
	ErrDeviceNotRegistered = errors.New("device not registered with the ether")
)

type packetInFlight struct {
	ReceivedPacket []byte
	ArrivalTime    time.Time
	Destination    *fake.Device
}

// TransformOutcome describes a packet that will be delivered to a receiver, created by a PacketTransform
type TransformOutcome struct {
	ReceivedPacket []byte
	FlightTime     time.Duration
}

// A PacketTransform is both a predicate and map function for packets transmitted through the Ether:
// - If the resulting outcome has a nil packet then no packet will be delivered to the receiver
// - Otherwise the resulting packet will be scheduled to arrive after a specified delay
//
// By providing custom PacketTransform functions, callers can add in logic to more accurately represent their
// transmission medium: adding in delays, corrupting the received packet, preventing receipt, etc.
type PacketTransform func(transmitter, receiver *fake.Device, packet []byte) TransformOutcome

// PerfectPacketTransform is a simple PacketTransform that always delivers all packets with no arrival delay
func PerfectPacketTransform(_, _ *fake.Device, packet []byte) TransformOutcome {
	return TransformOutcome{
		ReceivedPacket: packet,
	}
}

// AddPacketDeliveryDelay takes a PacketTransform and adds a delay to the resulting TransformOutcome
func AddPacketDeliveryDelay(transform PacketTransform, delay time.Duration) PacketTransform {
	return func(transmitter, receiver *fake.Device, packet []byte) (outcome TransformOutcome) {
		outcome = transform(transmitter, receiver, packet)
		outcome.FlightTime += delay
		return
	}
}

// Config configures the behaviour of the Ether
type Config struct {
	Logger          logr.Logger
	Clock           clockwork.Clock
	PacketTransform PacketTransform
}

// Ether allows for modelling a radio transmission medium when using fake.Device to simulate radio devices
type Ether struct {
	cfg             Config
	devices         map[*fake.Device]*connectedDevice
	packetsInFlight []*packetInFlight

	stop    chan struct{}
	actions chan func()
	stopped chan error
}

// New creates a new Ether
func New(cfg Config) *Ether {
	if cfg.Logger == nil {
		cfg.Logger = logr.Discard()
	}

	if cfg.Clock == nil {
		cfg.Clock = clockwork.NewRealClock()
	}

	if cfg.PacketTransform == nil {
		cfg.PacketTransform = PerfectPacketTransform
	}

	e := &Ether{
		cfg:     cfg,
		stop:    make(chan struct{}),
		stopped: make(chan error),
		actions: make(chan func()),
		devices: map[*fake.Device]*connectedDevice{},
	}

	go func() {
		defer close(e.stopped)
		e.stopped <- e.run()
	}()

	return e
}

func (e *Ether) run() error {
	e.cfg.Logger.Info("ether task started")
	defer e.cfg.Logger.Info("ether task exiting")

	for {
		var nextWake <-chan time.Time
		if len(e.packetsInFlight) > 0 {
			nextPacketToArrive := e.packetsInFlight[0]
			timeToWake := nextPacketToArrive.ArrivalTime.Sub(e.cfg.Clock.Now())
			nextWake = e.cfg.Clock.After(timeToWake)
		}

		select {
		case <-e.stop:
			return nil
		case action := <-e.actions:
			action()
		case <-nextWake:
			if len(e.packetsInFlight) < 1 {
				continue
			}

			nextPacket := e.packetsInFlight[0]
			e.packetsInFlight = e.packetsInFlight[1:]

			rxDeviceCtx := e.devices[nextPacket.Destination]
			if rxDeviceCtx == nil {
				continue
			}

			select {
			case rxDeviceCtx.rxChan <- nextPacket.ReceivedPacket:
			default:
			}
		}
	}
}

// Close gracefully shuts down the Ether and any background tasks
func (e *Ether) Close() error {
	var devices []*fake.Device
	_ = e.doSync(context.Background(), func() error {
		for d := range e.devices {
			devices = append(devices, d)
		}
		return nil
	})

	for _, d := range devices {
		e.UnregisterDevice(d)
	}

	close(e.stop)

	return <-e.stopped
}

type connectedDevice struct {
	Device *fake.Device
	rxChan chan []byte
}

// RegisterDevice connects a fake device to the ether so that its transmissions are propagated, and it may receive
// transmissions from other connected fake devices
func (e *Ether) RegisterDevice(device *fake.Device) {
	_ = e.doSync(context.Background(), func() error {
		device.Radio.Tx = e.radioTransmitIntoEther
		device.Radio.Rx = e.radioReceiveFromEther

		rxChan := make(chan []byte)
		e.devices[device] = &connectedDevice{
			Device: device,
			rxChan: rxChan,
		}

		return nil
	})
}

// UnregisterDevice removes a fake device from the ether, preventing further transmissions from reaching this device
func (e *Ether) UnregisterDevice(device *fake.Device) {
	_ = e.doSync(context.Background(), func() error {
		deviceContext := e.devices[device]
		if deviceContext == nil {
			return nil
		}

		close(deviceContext.rxChan)
		delete(e.devices, device)

		return nil
	})
}

func (e *Ether) radioTransmitIntoEther(d *fake.Device, packet []byte) error {
	return e.doSync(context.Background(), func() error {
		radioCtx := e.devices[d]
		if radioCtx == nil {
			return ErrDeviceNotRegistered
		}

		e.cfg.Logger.V(2).Info("device transmitting into ether", "device", addr(d))

		for otherDevice := range e.devices {
			if otherDevice == d {
				continue
			}

			transformOutcome := e.cfg.PacketTransform(d, otherDevice, packet)
			if transformOutcome.ReceivedPacket == nil {
				continue
			}

			e.cfg.Logger.V(3).Info("scheduling packet delivery", "sender", addr(d), "receiver", addr(otherDevice), "flight-time", transformOutcome.FlightTime)
			inFlight := &packetInFlight{
				ReceivedPacket: transformOutcome.ReceivedPacket,
				ArrivalTime:    e.cfg.Clock.Now().Add(transformOutcome.FlightTime),
				Destination:    otherDevice,
			}

			// sorted insert, ensuring the next packet to arrive is at the front
			insertIdx := sort.Search(len(e.packetsInFlight), func(i int) bool {
				return inFlight.ArrivalTime.After(e.packetsInFlight[i].ArrivalTime)
			})
			e.packetsInFlight = append(e.packetsInFlight, nil)
			copy(e.packetsInFlight[insertIdx+1:], e.packetsInFlight[insertIdx:])
			e.packetsInFlight[insertIdx] = inFlight
		}

		return nil
	})
}

func (e *Ether) radioReceiveFromEther(d *fake.Device) <-chan []byte {
	var rxChan <-chan []byte

	err := e.doSync(context.Background(), func() error {
		radioCtx := e.devices[d]
		if radioCtx == nil {
			return ErrDeviceNotRegistered
		}

		rxChan = radioCtx.rxChan

		return nil
	})
	if err != nil {
		return nil
	}

	e.cfg.Logger.V(2).Info("device listening to ether", "device", addr(d))
	return rxChan
}

func (e *Ether) doSync(ctx context.Context, f func() error) error {
	errChan := make(chan error)
	e.actions <- func() {
		errChan <- f()
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}

func addr(v interface{}) string {
	return fmt.Sprintf("%p", v)
}

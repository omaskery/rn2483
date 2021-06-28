package ether_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"github.com/jonboulle/clockwork"
	"github.com/poy/onpar"
	. "github.com/poy/onpar/expect"
	. "github.com/poy/onpar/matchers"

	"github.com/omaskery/rn2483"
	"github.com/omaskery/rn2483/fake"
	"github.com/omaskery/rn2483/fake/ether"
	"github.com/omaskery/rn2483/testutils"
)

type testDevice struct {
	name       string
	fakeDevice *fake.Device
	device     *rn2483.Device
}

type testContext struct {
	ctx       context.Context
	ctxCancel context.CancelFunc
	logger    logr.Logger
	devices   []*testDevice
	ether     *ether.Ether
	clock     clockwork.FakeClock
}

func prepareTestContext(t *testing.T) *testContext {
	ctx, cancel := context.WithCancel(context.Background())

	logger := testutils.CreateTestLogger(t)
	stdr.SetVerbosity(100)
	clock := clockwork.NewFakeClock()

	return &testContext{
		ctx:       ctx,
		ctxCancel: cancel,
		logger:    logger,
		clock:     clock,
		ether: ether.New(ether.Config{
			Logger: logger.WithName("ether"),
			Clock:  clock,
		}),
	}
}

func (t *testContext) AddDevice(name string) *testDevice {
	d := &testDevice{
		name: name,
	}
	d.fakeDevice, d.device = fake.NewFakeDevice(fake.Config{
		Logger: t.logger.WithName("fake-device").WithValues("device-name", name),
	})
	t.devices = append(t.devices, d)
	t.ether.RegisterDevice(d.fakeDevice)
	return d
}

func TestEther(t *testing.T) {
	o := onpar.New()
	defer o.Run(t)

	o.BeforeEach(func(t *testing.T) (*testing.T, *testContext) {
		return t, prepareTestContext(t)
	})

	o.AfterEach(func(t *testing.T, ctx *testContext) {
		if err := ctx.ether.Close(); err != nil {
			ctx.logger.Error(err, "error shutting down ether")
		}

		for i := len(ctx.devices) - 1; i >= 0; i-- {
			d := ctx.devices[i]
			if err := d.device.Close(); err != nil {
				ctx.logger.Error(err, "error shutting down fake device", "device-name", d.name)
			}
		}

		ctx.ctxCancel()
	})

	o.Spec("able to transmit through the ether", func(t *testing.T, ctx *testContext) {
		deviceA := ctx.AddDevice("device-a").device
		deviceB := ctx.AddDevice("device-b").device

		_, err := deviceA.PauseMAC()
		Expect(t, err).To(Not(HaveOccurred()))
		_, err = deviceB.PauseMAC()
		Expect(t, err).To(Not(HaveOccurred()))

		rxChan := make(chan []byte)
		go func() {
			defer close(rxChan)
			rx, err := deviceB.RadioRx(rn2483.ContinuousReceiveMode)
			Expect(t, err).To(Not(HaveOccurred()))
			rxChan <- rx
		}()

		// give the receiver time to spin up
		time.Sleep(10 * time.Millisecond)

		testData := []byte("hello world!")
		Expect(t, deviceA.RadioTx(testData)).To(Not(HaveOccurred()))

		select {
		case received := <-rxChan:
			Expect(t, received).To(Equal(testData))
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("timed out")
		}
	})
}

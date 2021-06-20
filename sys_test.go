package rn2483_test

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"github.com/poy/onpar"
	. "github.com/poy/onpar/expect"
	. "github.com/poy/onpar/matchers"

	"github.com/omaskery/rn2483"
	"github.com/omaskery/rn2483/fake"
)

type testContext struct {
	logger logr.Logger
	fake   *fake.Device
	device *rn2483.Device
}

func prepareTestContext(t *testing.T) *testContext {
	logger := createTestLogger(t)
	stdr.SetVerbosity(100)

	f := fake.New(fake.Config{
		Logger: logger.WithName("fake-device"),
	})
	d := rn2483.New(rn2483.Config{
		Serial: &rn2483.DebugSerial{
			Serial:     f,
			Logger:     logger.WithName("dbg-serial"),
			AssumeText: true,
		},
	})
	t.Cleanup(func() {
		if err := d.Close(); err != nil {
			logger.Error(err, "error cleaning up fake device")
		}
	})

	return &testContext{
		logger: logger,
		fake:   f,
		device: d,
	}
}

func TestSys(t *testing.T) {
	o := onpar.New()
	defer o.Run(t)

	o.BeforeEach(func(t *testing.T) (*testing.T, *testContext) {
		return t, prepareTestContext(t)
	})

	o.Spec("can retrieve & parse the device version", func(t *testing.T, ctx *testContext) {
		version, err := ctx.device.GetVersion()
		Expect(t, err).To(Not(HaveOccurred()))
		Expect(t, version.Raw).To(Equal("RN2483 1.0.4 Mar 23 1991 13:37:00"))
		Expect(t, version.VersionString()).To(Equal("1.0.4"))
		Expect(t, version.Major).To(Equal(1))
		Expect(t, version.Minor).To(Equal(0))
		Expect(t, version.Revision).To(Equal(4))
		Expect(t, version.ReleaseTime).To(Equal(time.Date(1991, time.March, 23, 13, 37, 00, 0, time.UTC)))
		Expect(t, version.IsKnownSKU()).To(BeTrue())
	})

	o.Spec("can reset the device", func(t *testing.T, ctx *testContext) {
		_, err := ctx.device.Reset()
		Expect(t, err).To(Not(HaveOccurred()))

		// TODO: actually assert that the fake has reverted some kind of non-persisted state
	})

	o.Spec("can read voltage", func(t *testing.T, ctx *testContext) {
		voltage, err := ctx.device.GetVDD()
		Expect(t, err).To(Not(HaveOccurred()))
		Expect(t, voltage.Volts()).To(And(BeAbove(3.290), BeBelow(3.310)))
		Expect(t, float64(voltage.Millivolts())).To(And(BeAbove(3290), BeBelow(3310)))
	})

	o.Spec("can write GPIO pins", func(t *testing.T, ctx *testContext) {
		for _, pin := range rn2483.AllPins {
			Expect(t, ctx.device.SetDigitalGPIO(pin, true)).To(Not(HaveOccurred()))
			Expect(t, ctx.fake.Sys.GPIO[pin]).To(BeTrue())

			Expect(t, ctx.device.SetDigitalGPIO(pin, false)).To(Not(HaveOccurred()))
			Expect(t, ctx.fake.Sys.GPIO[pin]).To(BeFalse())
		}
	})

	o.Group("non-volatile memory access", func() {
		o.Spec("can read NVM", func(t *testing.T, ctx *testContext) {
			data, err := rn2483.ReadNVM(ctx.device, rn2483.UserNVMStart, rn2483.UserNVMLength)
			Expect(t, err).To(Not(HaveOccurred()))

			expected := make([]byte, rn2483.UserNVMLength)
			for i := range expected {
				expected[i] = 0xFF
			}
			Expect(t, data).To(Equal(expected))
		})

		o.Spec("can write to NVM and read that data back", func(t *testing.T, ctx *testContext) {
			testData := []byte("hello, world!\r\n")

			err := rn2483.WriteNVM(ctx.device, rn2483.UserNVMStart, testData)
			Expect(t, err).To(Not(HaveOccurred()))

			data, err := rn2483.ReadNVM(ctx.device, rn2483.UserNVMStart, uint16(len(testData)))
			Expect(t, err).To(Not(HaveOccurred()))
			Expect(t, data).To(Equal(testData))
		})

		o.Group("reading & writing out of bounds fails", func() {
			invalidAddresses := []uint16{
				0, 0x2FF, 0x400, 0x500,
			}

			for _, address := range invalidAddresses {
				address := address
				o.Spec(fmt.Sprintf("address 0x%03X", address), func(t *testing.T, ctx *testContext) {
					_, err := ctx.device.ReadNVM(address)
					Expect(t, err).To(MatchError(rn2483.ErrInvalidParam))

					err = ctx.device.WriteNVM(address, 0x00)
					Expect(t, err).To(MatchError(rn2483.ErrInvalidParam))
				})
			}
		})
	})
}

type errorMatcher struct {
	expected error
}

func MatchError(err error) *errorMatcher {
	return &errorMatcher{
		expected: err,
	}
}

var _ Matcher = (*errorMatcher)(nil)

func (e errorMatcher) Match(actual interface{}) (resultValue interface{}, err error) {
	actualErr, ok := actual.(error)
	if !ok {
		return actual, fmt.Errorf("expected an error type, got type %T", actual)
	}

	if !errors.Is(actualErr, e.expected) {
		return actual, fmt.Errorf("expected error type %T, got type %T", e.expected, actualErr)
	}

	return actual, nil
}

func createTestLogger(t *testing.T) logr.Logger {
	buffer := &bytes.Buffer{}
	t.Cleanup(func() {
		if t.Failed() {
			fmt.Println("dumping log output for failed test:")
			fmt.Print(buffer)
		}
	})

	logger := stdr.New(log.New(buffer, "", log.LstdFlags)).WithName(t.Name())

	return logger
}

package rn2483_test

import (
	"testing"

	"github.com/poy/onpar"
	. "github.com/poy/onpar/expect"
	. "github.com/poy/onpar/matchers"

	"github.com/omaskery/rn2483"
	"github.com/omaskery/rn2483/fake"
)

func TestRadio(t *testing.T) {
	o := onpar.New()
	defer o.Run(t)

	o.BeforeEach(func(t *testing.T) (*testing.T, *testContext) {
		ctx := prepareTestContext(t)

		_, err := ctx.device.PauseMAC()
		Expect(t, err).To(Not(HaveOccurred()))

		return t, ctx
	})

	o.Spec("can set radio power", func(t *testing.T, ctx *testContext) {
		Expect(t, ctx.device.SetRadioPower(5)).To(Not(HaveOccurred()))
		power, err := ctx.device.GetRadioPower()
		Expect(t, err).To(Not(HaveOccurred()))
		Expect(t, power).To(Equal(5))
	})

	o.Spec("can transmit", func(t *testing.T, ctx *testContext) {
		testData := []byte("Hello, World!")

		var transmitted []byte
		ctx.fake.Radio.Tx = func(d *fake.Device, packet []byte) error {
			transmitted = packet
			return nil
		}

		err := ctx.device.RadioTx(testData)
		Expect(t, err).To(Not(HaveOccurred()))
		Expect(t, transmitted).To(Equal(testData))
	})

	o.Spec("can receive", func(t *testing.T, ctx *testContext) {
		testData := []byte("Wow, such test data!")

		rxChan := make(chan []byte)
		ctx.fake.Radio.Rx = func(d *fake.Device) <-chan []byte {
			return rxChan
		}

		go func() {
			rxChan <- testData
		}()

		data, err := ctx.device.RadioRx(rn2483.ContinuousReceiveMode)
		Expect(t, err).To(Not(HaveOccurred()))
		Expect(t, data).To(Equal(testData))
	})
}

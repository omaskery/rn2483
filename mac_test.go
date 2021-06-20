package rn2483_test

import (
	"testing"

	"github.com/poy/onpar"
	. "github.com/poy/onpar/expect"
	. "github.com/poy/onpar/matchers"

	"github.com/omaskery/rn2483/fake"
)

func TestMac(t *testing.T) {
	o := onpar.New()
	defer o.Run(t)

	o.BeforeEach(func(t *testing.T) (*testing.T, *testContext) {
		return t, prepareTestContext(t)
	})

	o.Spec("able to pause mac", func(t *testing.T, ctx *testContext) {
		Expect(t, ctx.fake.Mac.IsPaused()).To(BeFalse())

		duration, err := ctx.device.PauseMAC()
		Expect(t, err).To(Not(HaveOccurred()))
		Expect(t, duration).To(Equal(fake.MaxPauseDuration))
		Expect(t, ctx.fake.Mac.IsPaused()).To(BeTrue())
	})
}

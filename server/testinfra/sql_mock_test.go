package testinfra

import (
	"testing"
	"time"

	"github.com/fundwit/go-commons/types"
	. "github.com/onsi/gomega"
)

func TestAnyPastTime(t *testing.T) {
	RegisterTestingT(t)

	t.Run("AnyPastTime work as expected", func(t *testing.T) {
		v := AnyPastTime{Range: 3 * time.Second}
		dv, _ := types.CurrentTimestamp().Value()
		Expect(v.Match(dv)).To(BeTrue())

		dv, _ = types.Timestamp(time.Now().Add(-2 * time.Second)).Value()
		Expect(v.Match(dv)).To(BeTrue())

		dv, _ = types.Timestamp(time.Now().Add(-4 * time.Second)).Value()
		Expect(v.Match(dv)).To(BeFalse())
	})
}

func TestAnyId(t *testing.T) {
	RegisterTestingT(t)

	t.Run("AnyId work as expected", func(t *testing.T) {
		v := AnyId{}
		Expect(v.Match(int64(1))).To(BeTrue())

		Expect(v.Match(int64(0))).To(BeFalse())
		Expect(v.Match(int64(-1))).To(BeFalse())

		Expect(v.Match(int32(1))).To(BeFalse())
	})
}

func TestAnyArgument(t *testing.T) {
	RegisterTestingT(t)

	t.Run("AnyArgument work as expected", func(t *testing.T) {
		v := AnyArgument{}
		Expect(v.Match(1)).To(BeTrue())
		Expect(v.Match(nil)).To(BeTrue())
		Expect(v.Match(struct{}{})).To(BeTrue())
	})
}

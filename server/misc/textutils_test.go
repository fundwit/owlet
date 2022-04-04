package misc

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestSplit(t *testing.T) {
	RegisterTestingT(t)

	t.Run("Split should work as expected", func(t *testing.T) {
		Expect(Split("  aaa,  \t bbb,ccc ")).To(Equal([]string{"aaa", "bbb", "ccc"}))
		Expect(Split("   ")).To(Equal([]string{}))
		Expect(Split("")).To(Equal([]string{}))
	})
}

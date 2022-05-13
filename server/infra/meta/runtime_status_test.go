package meta

import (
	"runtime"
	"testing"

	. "github.com/onsi/gomega"
)

func TestAcquireRuntimeInfo(t *testing.T) {
	RegisterTestingT(t)

	t.Run("acquire runtime info successfully", func(t *testing.T) {
		r := AcquireRuntimeStatus()
		Expect(r).To(Equal(RuntimeStatus{
			StartTime:    startTime,
			NumCPU:       runtime.NumCPU(),
			NumGoroutine: runtime.NumGoroutine(),
			NumMaxProcs:  runtime.GOMAXPROCS(0),
		}))
	})
}

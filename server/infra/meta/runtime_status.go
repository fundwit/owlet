package meta

import (
	"runtime"

	"github.com/fundwit/go-commons/types"
)

type RuntimeStatus struct {
	StartTime types.Timestamp `json:"startTime"`

	NumCPU       int `json:"numCpu"`
	NumGoroutine int `json:"numGoroutine"`
	NumMaxProcs  int `json:"numMaxProcs"`
}

var (
	startTime                = types.CurrentTimestamp()
	AcquireRuntimeStatusFunc = AcquireRuntimeStatus
)

func AcquireRuntimeStatus() RuntimeStatus {
	return RuntimeStatus{
		StartTime:    startTime,
		NumCPU:       runtime.NumCPU(),
		NumGoroutine: runtime.NumGoroutine(),
		NumMaxProcs:  runtime.GOMAXPROCS(0),
	}
}

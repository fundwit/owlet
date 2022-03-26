package meta

import (
	"fmt"
	"owlet/server/infra/idgen"
	"runtime"
	"time"

	"github.com/sony/sonyflake"
)

type ServiceMeta struct {
	Name       string    `json:"name"`
	InstanceID string    `json:"instanceId"`
	StartTime  time.Time `json:"startTime"`
}

type ServiceInfo struct {
	ServiceMeta

	Duration int64 `json:"duration"`

	NumCPU       int `json:"numCpu"`
	NumGoroutine int `json:"numGoroutine"`
	NumMaxProcs  int `json:"numMaxProcs"`
}

var idWorker = sonyflake.NewSonyflake(sonyflake.Settings{})
var serviceMeta = ServiceMeta{
	Name:       "owlet",
	InstanceID: fmt.Sprint(idgen.NextID(idWorker)),
	StartTime:  time.Now(),
}

func GetServiceMeta() ServiceInfo {
	return ServiceInfo{
		ServiceMeta:  serviceMeta,
		Duration:     time.Now().Unix() - serviceMeta.StartTime.Unix(),
		NumCPU:       runtime.NumCPU(),
		NumGoroutine: runtime.NumGoroutine(),
		NumMaxProcs:  runtime.GOMAXPROCS(0),
	}
}

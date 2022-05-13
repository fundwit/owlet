package meta

import (
	"fmt"
	"owlet/server/infra/idgen"

	"github.com/sony/sonyflake"
)

type ServiceIdentity struct {
	Name       string `json:"name"`
	InstanceID string `json:"instanceId"`
}

var idWorker = sonyflake.NewSonyflake(sonyflake.Settings{})
var serviceInstanceId = fmt.Sprint(idgen.NextID(idWorker))
var AcquireServiceIdentityFunc = AcquireServiceIdentity

func AcquireServiceIdentity() ServiceIdentity {
	return ServiceIdentity{
		Name:       "owlet",
		InstanceID: serviceInstanceId,
	}
}

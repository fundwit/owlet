package idgen

import (
	"github.com/fundwit/go-commons/types"
	"github.com/sony/sonyflake"
)

func NextID(idWorker *sonyflake.Sonyflake) types.ID {
	// After the Sonyflake time overflows (sf.elapsedTime >= 1<<BitLenTime),
	// NextID returns an errors.New("over the time limit")
	id, err := idWorker.NextID()
	if err != nil {
		panic(err)
	}
	return types.ID(id)
}

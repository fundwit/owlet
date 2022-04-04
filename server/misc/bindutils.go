package misc

import (
	"owlet/server/infra/fail"
	"strconv"

	"github.com/fundwit/go-commons/types"
	"github.com/gin-gonic/gin"
)

type requestPath struct {
	ID types.ID `uri:"id" binding:"required"`
}

type UidObject struct {
	UID types.ID `json:"uid"`
}

type IdObject struct {
	ID types.ID `json:"id"`
}

func NewIdObject(id types.ID) *IdObject {
	return &IdObject{ID: id}
}

func BindingPathID(c *gin.Context) (types.ID, error) {
	p := requestPath{}
	if err := c.ShouldBindUri(&p); err != nil {
		// maybe: strconv.NumError{Func, Num, Err: strconv.ErrRange|strconv.ErrSyntax|...}
		if d, ok := err.(*strconv.NumError); ok {
			return 0, &fail.ErrBadParam{Param: "id", InvalidValue: d.Num, Cause: err}
		}
		return 0, &fail.ErrBadParam{Cause: err}
	}
	return p.ID, nil
}

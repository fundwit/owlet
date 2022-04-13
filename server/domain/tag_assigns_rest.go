package domain

import (
	"errors"
	"net/http"
	"owlet/server/infra/fail"
	"owlet/server/infra/sessions"

	"github.com/gin-gonic/gin"
)

var (
	PathTagAssigns = "/v1/tag-assigns"
)

func RegisterTagAssignsRestAPI(r *gin.Engine, middleWares ...gin.HandlerFunc) {
	g := r.Group(PathTagAssigns, middleWares...)
	g.POST("", handleCreateTagAssigns)
	g.DELETE("", handleDeleteTagAssignWithQuery)
}

// @ID tag-assign-create
// @Param _ body domain.TagAssignCreate true "request body"
// @Success 201 {object} misc.IdObject
// @Failure default {object} fail.ErrorBody "error"
// @Router /v1/tag-assigns/ [post]
func handleCreateTagAssigns(c *gin.Context) {
	p := TagAssignCreate{}
	err := c.ShouldBindJSON(&p)
	if err != nil && err.Error() == "EOF" {
		panic(&fail.ErrBadParam{Cause: errors.New("empty body")})
	}
	if err != nil {
		panic(&fail.ErrBadParam{Cause: err})
	}

	resp, err := CreateTagAssignFunc(&p, sessions.ExtractSessionFromGinContext(c))
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, resp)
}

// @ID tag-assign-query-delete
// @Param resId query int true "resource id"
// @Param tagId query int true "tag id"
// @Success 204 {object} string "response body is empty"
// @Failure default {object} fail.ErrorBody "error"
// @Router /v1/tag-assigns [delete]
func handleDeleteTagAssignWithQuery(c *gin.Context) {
	q := TagAssignRelation{}
	err := c.ShouldBindQuery(&q)
	if err != nil {
		panic(&fail.ErrBadParam{Cause: err})
	}

	err = DeleteTagAssignWithQueryFunc(&q, sessions.ExtractSessionFromGinContext(c))
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusNoContent, gin.H{})
}

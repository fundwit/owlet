package domain

import (
	"net/http"
	"owlet/server/infra/sessions"

	"github.com/gin-gonic/gin"
)

var (
	PathTags = "/v1/tags"
)

func RegisterTagsRestAPI(r *gin.Engine, middleWares ...gin.HandlerFunc) {
	g := r.Group(PathTags, middleWares...)
	g.GET("", handleQueryTags)
}

// @ID tag-with-stat-list
// @Success 200 {array} domain.TagWithStat
// @Failure default {object} fail.ErrorBody "error"
// @Router /v1/tags [get]
func handleQueryTags(c *gin.Context) {
	record, err := QueryTagsWithStatFunc(sessions.ExtractSessionFromGinContext(c))
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, record)
}

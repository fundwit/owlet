package domain

import (
	"net/http"
	"owlet/server/infra/fail"
	"owlet/server/infra/sessions"
	"owlet/server/misc"

	"github.com/gin-gonic/gin"
)

var (
	PathArticles = "/v1/articles"
)

func RegisterArticlesRestAPI(r *gin.Engine, middleWares ...gin.HandlerFunc) {
	g := r.Group(PathArticles, middleWares...)
	g.GET("", handleQueryArticles)
	g.GET(":id", handleDetailArticle)
}

// @ID article-meta-list
// @Param kw query string false "query keyword"
// @Param page query int false "page number based 1"
// @Success 200 {array} domain.ArticleMetaExt
// @Failure default {object} fail.ErrorBody "error"
// @Router /v1/articles [get]
func handleQueryArticles(c *gin.Context) {
	q := ArticleQuery{}
	err := c.ShouldBindQuery(&q)
	if err != nil {
		panic(&fail.ErrBadParam{Cause: err})
	}

	record, err := QueryArticlesFunc(q, &sessions.Session{Context: c.Request.Context(), Identity: sessions.Identity{ID: 1}})
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, record)
}

// @ID article-detail
// @Param id path uint64 true "id"
// @Success 200 {object} domain.ArticleDetail "response body"
// @Failure default {object} fail.ErrorBody "error"
// @Router /v1/articles/{id} [get]
func handleDetailArticle(c *gin.Context) {
	id, err := misc.BindingPathID(c)
	if err != nil {
		panic(err)
	}

	detail, err := DetailArticleFunc(id, &sessions.Session{Context: c.Request.Context()})
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, detail)
}

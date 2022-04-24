package domain

import (
	"errors"
	"net/http"
	"owlet/server/infra/fail"
	"owlet/server/infra/sessions"
	"owlet/server/misc"

	"github.com/fundwit/go-commons/types"
	"github.com/gin-gonic/gin"
)

var (
	PathArticles = "/v1/articles"
)

func RegisterArticlesRestAPI(r *gin.Engine, middleWares ...gin.HandlerFunc) {
	g := r.Group(PathArticles, middleWares...)
	g.GET("", handleQueryArticles)
	g.POST("", handleCreateArticle)
	g.GET(":id", handleDetailArticle)
	g.PUT(":id", handlePatchArticle)
	g.DELETE(":id", handleDeleteArticle)
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

	record, err := QueryArticlesFunc(q, sessions.ExtractSessionFromGinContext(c))
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, record)
}

// @ID article-create
// @Param _ body domain.ArticleCreate true "request body"
// @Success 201 {object} misc.IdObject
// @Failure default {object} fail.ErrorBody "error"
// @Router /v1/articles/ [post]
func handleCreateArticle(c *gin.Context) {
	p := ArticleCreate{}
	err := c.ShouldBindJSON(&p)
	if err != nil && err.Error() == "EOF" {
		panic(&fail.ErrBadParam{Cause: errors.New("empty body")})
	}
	if err != nil {
		panic(&fail.ErrBadParam{Cause: err})
	}

	id, err := CreateArticleFunc(&p, sessions.ExtractSessionFromGinContext(c))
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, misc.NewIdObject(id))
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

	detail, err := DetailArticleFunc(id, sessions.ExtractSessionFromGinContext(c))
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, detail)
}

type PatchResponse struct {
	ID         types.ID        `json:"id"`
	ModifyTime types.Timestamp `json:"modifyTime"`
}

// @ID article-patch
// @Param id path uint64 true "id"
// @Param _ body domain.ArticlePatch true "request body"
// @Success 200 {object} string "response body is empty"
// @Failure default {object} fail.ErrorBody "error"
// @Router /v1/articles/{id} [put]
func handlePatchArticle(c *gin.Context) {
	id, err := misc.BindingPathID(c)
	if err != nil {
		panic(err)
	}

	p := ArticlePatch{}
	err = c.ShouldBindJSON(&p)
	if err != nil && err.Error() == "EOF" {
		panic(&fail.ErrBadParam{Cause: errors.New("empty body")})
	}
	if err != nil {
		panic(&fail.ErrBadParam{Cause: err})
	}

	rt, err := PatchArticleFunc(id, &p, sessions.ExtractSessionFromGinContext(c))
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, &PatchResponse{ID: id, ModifyTime: *rt})
}

// @ID article-delete
// @Param id path uint64 true "id"
// @Success 204 {object} string "response body is empty"
// @Failure default {object} fail.ErrorBody "error"
// @Router /v1/articles/{id} [delete]
func handleDeleteArticle(c *gin.Context) {
	id, err := misc.BindingPathID(c)
	if err != nil {
		panic(err)
	}

	err = DeleteArticleFunc(id, sessions.ExtractSessionFromGinContext(c))
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusNoContent, gin.H{})
}

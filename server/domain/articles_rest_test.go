package domain

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"owlet/server/infra/fail"
	"owlet/server/infra/sessions"
	"owlet/server/testinfra"
	"testing"
	"time"

	"github.com/fundwit/go-commons/types"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
)

func TestQueryArticlesAPI(t *testing.T) {
	RegisterTestingT(t)

	router := gin.Default()
	router.Use(fail.ErrorHandling())
	RegisterArticlesRestAPI(router)

	t.Run("should be able to handle error on query articles", func(t *testing.T) {
		QueryArticlesFunc = func(q ArticleQuery, s *sessions.Session) ([]ArticleMetaExt, error) {
			return nil, errors.New("some error")
		}

		req := httptest.NewRequest(http.MethodGet, PathArticles, nil)
		status, body, _ := testinfra.ExecuteRequest(req, router)
		Expect(body).To(MatchJSON(`{"code":"common.internal_server_error", "message":"some error", "data":null}`))
		Expect(status).To(Equal(http.StatusInternalServerError))
	})

	t.Run("should be able to handle error on binding", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, PathArticles+"?kw="+testinfra.Alphabeta100+testinfra.Alphabeta100+"1", nil)
		status, body, _ := testinfra.ExecuteRequest(req, router)
		Expect(body).To(MatchJSON(`{"code":"common.bad_param",
			"message":"Key: 'ArticleQuery.KeyWord' Error:Field validation for 'KeyWord' failed on the 'lte' tag",
			"data":null}`))
		Expect(status).To(Equal(http.StatusBadRequest))
	})

	t.Run("should be able to handle query request successfully", func(t *testing.T) {
		var in ArticleQuery
		QueryArticlesFunc = func(q ArticleQuery, s *sessions.Session) ([]ArticleMetaExt, error) {
			in = q
			am := ArticleMeta{
				ID: 100, Type: GenericTypeIT, Title: "demo article", UID: 10,
				CreateTime: types.TimestampOfDate(2022, 1, 2, 3, 4, 5, 0, time.UTC),
				ModifyTime: types.TimestampOfDate(2022, 1, 2, 3, 4, 5, 0, time.UTC),
				Status:     ArticleStatusPublished, IsInvalid: false, Abstracts: "demo",
				Source: ArticleSourceOriginal, IsElite: true, IsTop: true, ViewNum: 30, CommentNum: 20,
			}
			return []ArticleMetaExt{
				{ArticleMeta: am, Tags: []Tag{{ID: 1000, Name: "go", Image: "go.png", Note: "golang"}}},
			}, nil
		}

		req := httptest.NewRequest(http.MethodGet, PathArticles+"?kw=demo&page=2", nil)
		status, body, _ := testinfra.ExecuteRequest(req, router)
		Expect(status).To(Equal(http.StatusOK))
		Expect(body).To(MatchJSON(`[{"id": "100", "type": 2, "title": "demo article", "uid": "10",
			"create_time": "2022-01-02T03:04:05Z", "modify_time": "2022-01-02T03:04:05Z",
			"status": 1, "is_invalid": false, "abstracts": "demo", "source": 1,
			"is_elite": true, "is_top": true, "view_num": 30, "comment_num": 20,
			"tags": [{"id": "1000", "name":"go", "image":"go.png", "note": "golang"}]}]`))

		Expect(in).To(Equal(ArticleQuery{KeyWord: "demo", Page: 2}))
	})
}

func TestDetailArticlesAPI(t *testing.T) {
	RegisterTestingT(t)

	router := gin.Default()
	router.Use(fail.ErrorHandling())
	RegisterArticlesRestAPI(router)

	t.Run("should be able to get article detail", func(t *testing.T) {
		meta := ArticleMeta{
			ID: 100, Type: GenericTypeIT, Title: "demo article", UID: 10,
			CreateTime: types.TimestampOfDate(2022, 1, 2, 3, 4, 5, 0, time.UTC),
			ModifyTime: types.TimestampOfDate(2022, 1, 2, 3, 4, 5, 0, time.UTC),
			Status:     ArticleStatusPublished, IsInvalid: false, Abstracts: "demo",
			Source: ArticleSourceOriginal, IsElite: true, IsTop: true, ViewNum: 30, CommentNum: 20,
		}
		tags := []Tag{{ID: 1000, Name: "go", Image: "go.png", Note: "golang"}}

		DetailArticleFunc = func(id types.ID, s *sessions.Session) (*ArticleDetail, error) {
			Expect(id).To(Equal(types.ID(200)))
			return &ArticleDetail{ArticleRecord: ArticleRecord{ArticleMeta: meta, Content: "content 100"}, Tags: tags}, nil
		}

		req := httptest.NewRequest(http.MethodGet, PathArticles+"/200", nil)
		status, body, _ := testinfra.ExecuteRequest(req, router)

		Expect(status).To(Equal(http.StatusOK))
		Expect(body).To(MatchJSON(`{"id": "100", "type": 2, "title": "demo article", "uid": "10",
			"create_time": "2022-01-02T03:04:05Z", "modify_time": "2022-01-02T03:04:05Z", "status": 1,
			"is_invalid": false, "abstracts": "demo", "source": 1, "is_elite": true, "is_top": true,
			"view_num": 30, "comment_num": 20, "content": "content 100",
			"tags": [{"id": "1000", "name":"go", "image":"go.png", "note": "golang"}]
			}`))
	})

	t.Run("should be able to handle error on detail article", func(t *testing.T) {
		DetailArticleFunc = func(id types.ID, s *sessions.Session) (*ArticleDetail, error) {
			return nil, errors.New("some error")
		}

		req := httptest.NewRequest(http.MethodGet, PathArticles+"/100", nil)
		status, body, _ := testinfra.ExecuteRequest(req, router)
		Expect(body).To(MatchJSON(`{"code":"common.internal_server_error", "message":"some error", "data":null}`))
		Expect(status).To(Equal(http.StatusInternalServerError))
	})

	t.Run("should be able to handle error on bad params", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, PathArticles+"/abc", nil)
		status, body, _ := testinfra.ExecuteRequest(req, router)
		Expect(body).To(MatchJSON(`{"code":"common.bad_param", "message":"invalid id 'abc'", "data":null}`))
		Expect(status).To(Equal(http.StatusBadRequest))
	})
}

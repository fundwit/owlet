package domain

import (
	"bytes"
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
			Expect(*s).To(Equal(sessions.GuestSession))
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
			Expect(*s).To(Equal(sessions.GuestSession))
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

func TestCreateArticlesAPI(t *testing.T) {
	RegisterTestingT(t)

	router := gin.Default()
	router.Use(fail.ErrorHandling())
	RegisterArticlesRestAPI(router)

	t.Run("should be able to handle error on empty body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, PathArticles, nil)
		status, body, _ := testinfra.ExecuteRequest(req, router)
		Expect(body).To(MatchJSON(`{"code": "common.bad_param",
			"message": "empty body",
			"data":null}`))
		Expect(status).To(Equal(http.StatusBadRequest))
	})

	t.Run("should be able to handle error on binding failed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, PathArticles, bytes.NewReader([]byte(
			`{"type": 100, "status":200, "source": 300}`)))
		status, body, _ := testinfra.ExecuteRequest(req, router)
		Expect(body).To(MatchJSON(`{"code": "common.bad_param",
			"message": "Key: 'ArticleCreate.Title' Error:Field validation for 'Title' failed on the 'required' tag\n` +
			`Key: 'ArticleCreate.Content' Error:Field validation for 'Content' failed on the 'required' tag\n` +
			`Key: 'ArticleCreate.Type' Error:Field validation for 'Type' failed on the 'oneof' tag\n` +
			`Key: 'ArticleCreate.Source' Error:Field validation for 'Source' failed on the 'oneof' tag\n` +
			`Key: 'ArticleCreate.Status' Error:Field validation for 'Status' failed on the 'oneof' tag",
			"data":null}`))
		Expect(status).To(Equal(http.StatusBadRequest))
	})

	t.Run("should be able to handle error on query articles", func(t *testing.T) {
		CreateArticleFunc = func(q *ArticleCreate, s *sessions.Session) (types.ID, error) {
			Expect(*s).To(Equal(sessions.GuestSession))
			return 0, errors.New("some error")
		}
		req := httptest.NewRequest(http.MethodPost, PathArticles, bytes.NewReader([]byte(
			`{"title": "test title", "content": "test content", "type":1, "source":1, "status":1}`)))
		status, body, _ := testinfra.ExecuteRequest(req, router)
		Expect(body).To(MatchJSON(`{"code":"common.internal_server_error", "message":"some error", "data":null}`))
		Expect(status).To(Equal(http.StatusInternalServerError))
	})

	t.Run("should be able to handle query success", func(t *testing.T) {
		CreateArticleFunc = func(q *ArticleCreate, s *sessions.Session) (types.ID, error) {
			Expect(*s).To(Equal(sessions.GuestSession))
			Expect(q.Content).To(Equal("test content"))
			Expect(q.Title).To(Equal("test title"))
			Expect(q.Type).To(Equal(GenericType(2)))
			Expect(q.Source).To(Equal(ArticleSource(3)))
			Expect(q.Status).To(Equal(ArticleStatus(1)))
			Expect(q.IsTop).To(BeTrue())
			Expect(q.IsElite).To(BeTrue())
			return 100, nil
		}

		req := httptest.NewRequest(http.MethodPost, PathArticles, bytes.NewReader([]byte(
			`{"content": "test content", "title":"test title", "type":2, "source":3, "status":1, "is_top": true, "is_elite": true}`)))
		status, body, _ := testinfra.ExecuteRequest(req, router)
		Expect(body).To(MatchJSON(`{"id": "100"}`))
		Expect(status).To(Equal(http.StatusOK))
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
			Expect(*s).To(Equal(sessions.GuestSession))
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
			Expect(*s).To(Equal(sessions.GuestSession))
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

func TestPatchArticlesAPI(t *testing.T) {
	RegisterTestingT(t)

	router := gin.Default()
	router.Use(fail.ErrorHandling())
	RegisterArticlesRestAPI(router)

	t.Run("should be able to handle error on binding id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, PathArticles+"/abc", nil)
		status, body, _ := testinfra.ExecuteRequest(req, router)
		Expect(body).To(MatchJSON(`{"code": "common.bad_param",
			"message": "invalid id 'abc'",
			"data":null}`))
		Expect(status).To(Equal(http.StatusBadRequest))
	})

	t.Run("should be able to handle error on empty body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, PathArticles+"/100", nil)
		status, body, _ := testinfra.ExecuteRequest(req, router)
		Expect(body).To(MatchJSON(`{"code": "common.bad_param",
			"message": "empty body",
			"data":null}`))
		Expect(status).To(Equal(http.StatusBadRequest))
	})

	t.Run("should be able to handle error on binding failed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, PathArticles+"/100", bytes.NewReader([]byte(
			`{"content": "test content", "title": "test title", "type": 100, "status":200, "source": 300}`)))
		status, body, _ := testinfra.ExecuteRequest(req, router)
		Expect(body).To(MatchJSON(`{"code": "common.bad_param",
			"message": "Key: 'ArticlePatch.Type' Error:Field validation for 'Type' failed on the 'oneof' tag\n` +
			`Key: 'ArticlePatch.Status' Error:Field validation for 'Status' failed on the 'oneof' tag\n` +
			`Key: 'ArticlePatch.Source' Error:Field validation for 'Source' failed on the 'oneof' tag",
			"data":null}`))
		Expect(status).To(Equal(http.StatusBadRequest))
	})

	t.Run("should be able to handle error on query articles", func(t *testing.T) {
		PatchArticleFunc = func(id types.ID, p *ArticlePatch, s *sessions.Session) (*types.Timestamp, error) {
			Expect(*s).To(Equal(sessions.GuestSession))
			return nil, errors.New("some error")
		}
		req := httptest.NewRequest(http.MethodPut, PathArticles+"/100", bytes.NewReader([]byte(
			`{"type":1, "source":1, "status":1}`)))
		status, body, _ := testinfra.ExecuteRequest(req, router)
		Expect(body).To(MatchJSON(`{"code":"common.internal_server_error", "message":"some error", "data":null}`))
		Expect(status).To(Equal(http.StatusInternalServerError))
	})

	t.Run("should be able to handle query success", func(t *testing.T) {
		ts := types.TimestampOfDate(2022, 4, 5, 6, 10, 20, 0, time.UTC)
		PatchArticleFunc = func(id types.ID, p *ArticlePatch, s *sessions.Session) (*types.Timestamp, error) {
			Expect(*s).To(Equal(sessions.GuestSession))
			Expect(id).To(Equal(types.ID(100)))
			ip := *p
			Expect(ip.Content).To(Equal("test content"))
			Expect(ip.Title).To(Equal("test title"))
			Expect(*ip.Type).To(Equal(GenericType(2)))
			Expect(*ip.Source).To(Equal(ArticleSource(3)))
			Expect(*ip.Status).To(Equal(ArticleStatus(1)))
			Expect(*ip.IsTop).To(BeTrue())
			Expect(*ip.IsElite).To(BeTrue())
			Expect(ip.BaseModifyTime).To(Equal(types.TimestampOfDate(2022, 4, 5, 6, 7, 8, 0, time.UTC)))
			return &ts, nil
		}

		req := httptest.NewRequest(http.MethodPut, PathArticles+"/100", bytes.NewReader([]byte(
			`{"content": "test content", "title":"test title", "type":2, "source":3, "status":1, "is_top": true, "is_elite": true,
			  "baseModifyTime": "2022-04-05T06:07:08Z"}`)))
		status, body, _ := testinfra.ExecuteRequest(req, router)
		Expect(body).To(MatchJSON(`{"id": "100", "modifyTime": "2022-04-05T06:10:20Z"}`))
		Expect(status).To(Equal(http.StatusOK))
	})
}

func TestDeleteArticlesAPI(t *testing.T) {
	RegisterTestingT(t)

	router := gin.Default()
	router.Use(fail.ErrorHandling())
	RegisterArticlesRestAPI(router)

	t.Run("should be able to delete article", func(t *testing.T) {
		DeleteArticleFunc = func(id types.ID, s *sessions.Session) error {
			Expect(*s).To(Equal(sessions.GuestSession))
			Expect(id).To(Equal(types.ID(200)))
			return nil
		}

		req := httptest.NewRequest(http.MethodDelete, PathArticles+"/200", nil)
		status, body, _ := testinfra.ExecuteRequest(req, router)

		Expect(status).To(Equal(http.StatusNoContent))
		Expect(body).To(Equal(""))
	})

	t.Run("should be able to handle error on delete article", func(t *testing.T) {
		DeleteArticleFunc = func(id types.ID, s *sessions.Session) error {
			Expect(*s).To(Equal(sessions.GuestSession))
			Expect(id).To(Equal(types.ID(200)))
			return errors.New("some error")
		}

		req := httptest.NewRequest(http.MethodDelete, PathArticles+"/200", nil)
		status, body, _ := testinfra.ExecuteRequest(req, router)
		Expect(body).To(MatchJSON(`{"code":"common.internal_server_error", "message":"some error", "data":null}`))
		Expect(status).To(Equal(http.StatusInternalServerError))
	})

	t.Run("should be able to handle error on bad params", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, PathArticles+"/abc", nil)
		status, body, _ := testinfra.ExecuteRequest(req, router)
		Expect(body).To(MatchJSON(`{"code":"common.bad_param", "message":"invalid id 'abc'", "data":null}`))
		Expect(status).To(Equal(http.StatusBadRequest))
	})
}

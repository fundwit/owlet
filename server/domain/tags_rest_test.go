package domain

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"owlet/server/infra/fail"
	"owlet/server/infra/sessions"
	"owlet/server/testinfra"
	"testing"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
)

func TestQueryTagsAPI(t *testing.T) {
	RegisterTestingT(t)

	router := gin.Default()
	router.Use(fail.ErrorHandling())
	RegisterTagsRestAPI(router)

	t.Run("should be able to handle error", func(t *testing.T) {
		QueryTagsWithStatFunc = func(s *sessions.Session) ([]TagWithStat, error) {
			Expect(*s).To(Equal(sessions.GuestSession))
			return nil, errors.New("some error")
		}
		req := httptest.NewRequest(http.MethodGet, PathTags, nil)
		status, body, _ := testinfra.ExecuteRequest(req, router)
		Expect(status).To(Equal(http.StatusInternalServerError))
		Expect(body).To(MatchJSON(`{"code":"common.internal_server_error", "message":"some error", "data":null}`))
	})

	t.Run("should be able to handle query request successfully", func(t *testing.T) {
		QueryTagsWithStatFunc = func(s *sessions.Session) ([]TagWithStat, error) {
			Expect(*s).To(Equal(sessions.GuestSession))
			return []TagWithStat{
				{Tag: Tag{ID: 100, Name: "golang", Note: "go language", Image: "golang.png"}, Count: 10},
			}, nil
		}
		req := httptest.NewRequest(http.MethodGet, PathTags, nil)
		status, body, _ := testinfra.ExecuteRequest(req, router)
		Expect(status).To(Equal(http.StatusOK))
		Expect(body).To(MatchJSON(`[{"id": "100", "name": "golang", "note": "go language", "image": "golang.png", "count": 10}]`))
	})
}

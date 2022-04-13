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

	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
)

func TestCreateTagAssignAPI(t *testing.T) {
	RegisterTestingT(t)

	router := gin.Default()
	router.Use(fail.ErrorHandling())
	RegisterTagAssignsRestAPI(router)

	t.Run("should be able to handle error on empty body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, PathTagAssigns, nil)
		status, body, _ := testinfra.ExecuteRequest(req, router)
		Expect(body).To(MatchJSON(`{"code": "common.bad_param",
			"message": "empty body",
			"data":null}`))
		Expect(status).To(Equal(http.StatusBadRequest))
	})

	t.Run("should be able to handle error on binding failed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, PathTagAssigns, bytes.NewReader([]byte(
			`{}`)))
		status, body, _ := testinfra.ExecuteRequest(req, router)
		Expect(body).To(MatchJSON(`{"code": "common.bad_param",
			"message": "Key: 'TagAssignCreate.ResID' Error:Field validation for 'ResID' failed on the 'required' tag\n` +
			`Key: 'TagAssignCreate.TagName' Error:Field validation for 'TagName' failed on the 'required' tag",
			"data":null}`))
		Expect(status).To(Equal(http.StatusBadRequest))
	})

	t.Run("should be able to handle error on query articles", func(t *testing.T) {
		CreateTagAssignFunc = func(c *TagAssignCreate, s *sessions.Session) (*TagAssignCreateResponse, error) {
			return nil, errors.New("some error")
		}
		req := httptest.NewRequest(http.MethodPost, PathTagAssigns, bytes.NewReader([]byte(
			`{"resId": "100", "tagName": "test"}`)))
		status, body, _ := testinfra.ExecuteRequest(req, router)
		Expect(body).To(MatchJSON(`{"code":"common.internal_server_error", "message":"some error", "data":null}`))
		Expect(status).To(Equal(http.StatusInternalServerError))
	})

	t.Run("should be able to handle query success", func(t *testing.T) {
		CreateTagAssignFunc = func(c *TagAssignCreate, s *sessions.Session) (*TagAssignCreateResponse, error) {
			Expect(*c).To(Equal(TagAssignCreate{ResID: 300, TagName: "test"}))
			Expect(*s).To(Equal(sessions.GuestSession))
			return &TagAssignCreateResponse{
				TagAssignment: TagAssignment{
					ID: 1000, ResID: 300, ResType: 0, TagID: 2000, TagOrder: 0,
				},
				TagName: "test", TagNote: "Test", TagImage: "test.png",
			}, nil
		}

		req := httptest.NewRequest(http.MethodPost, PathTagAssigns, bytes.NewReader([]byte(
			`{"resId": "300", "tagName":"test"}`)))
		status, body, _ := testinfra.ExecuteRequest(req, router)
		Expect(body).To(MatchJSON(`{"id": "1000", "resId": "300", "tagId":"2000", "resType": 0, "tagOrder": 0,
			"tagName":"test", "tagNote": "Test", "tagImage":"test.png"}`))
		Expect(status).To(Equal(http.StatusOK))
	})
}

func TestDeleteTagAssignWithQueryAPI(t *testing.T) {
	RegisterTestingT(t)

	router := gin.Default()
	router.Use(fail.ErrorHandling())
	RegisterTagAssignsRestAPI(router)

	t.Run("should be able to handle error on binding failed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, PathTagAssigns, nil)
		status, body, _ := testinfra.ExecuteRequest(req, router)
		Expect(body).To(MatchJSON(`{"code": "common.bad_param",
			"message": "Key: 'TagAssignRelation.ResID' Error:Field validation for 'ResID' failed on the 'required' tag\n` +
			`Key: 'TagAssignRelation.TagID' Error:Field validation for 'TagID' failed on the 'required' tag",
			"data":null}`))
		Expect(status).To(Equal(http.StatusBadRequest))
	})

	t.Run("should be able to handle error on delete tag assigns with query", func(t *testing.T) {
		DeleteTagAssignWithQueryFunc = func(c *TagAssignRelation, s *sessions.Session) error {
			return errors.New("some error")
		}
		req := httptest.NewRequest(http.MethodDelete, PathTagAssigns+"?resId=100&tagId=200", nil)
		status, body, _ := testinfra.ExecuteRequest(req, router)
		Expect(body).To(MatchJSON(`{"code":"common.internal_server_error", "message":"some error", "data":null}`))
		Expect(status).To(Equal(http.StatusInternalServerError))
	})

	t.Run("should be able to handle delete tag assigns with query", func(t *testing.T) {
		DeleteTagAssignWithQueryFunc = func(c *TagAssignRelation, s *sessions.Session) error {
			Expect(*c).To(Equal(TagAssignRelation{ResID: 300, TagID: 400}))
			Expect(*s).To(Equal(sessions.GuestSession))
			return nil
		}

		req := httptest.NewRequest(http.MethodDelete, PathTagAssigns+"?resId=300&tagId=400", nil)
		status, body, _ := testinfra.ExecuteRequest(req, router)
		Expect(body).To(Equal(""))
		Expect(status).To(Equal(http.StatusNoContent))
	})
}

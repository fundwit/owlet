package sessions_test

import (
	"net/http"
	"net/http/httptest"
	"owlet/server/infra/fail"
	"owlet/server/infra/sessions"
	"owlet/server/testinfra"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
)

func TestSessionFilter(t *testing.T) {
	RegisterTestingT(t)

	engine := gin.Default()
	engine.Use(fail.ErrorHandling(), sessions.SessionFilter())
	engine.GET("/", func(c *gin.Context) {
		s := sessions.ExtractSessionFromGinContext(c)
		c.String(http.StatusOK, s.Token)
	})

	t.Run("unauthenticated response when token is absent", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		status, body, _ := testinfra.ExecuteRequest(req, engine)
		Expect(status).To(Equal(http.StatusUnauthorized))
		Expect(body).To(MatchJSON(`{"code":"security.unauthenticated", "message": "unauthenticated", "data": null}`))
	})

	t.Run("unauthenticated response when token is invalid", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Add("cookie", "sec_token=absent")
		status, body, _ := testinfra.ExecuteRequest(req, engine)
		Expect(status).To(Equal(http.StatusUnauthorized))
		Expect(body).To(MatchJSON(`{"code":"security.unauthenticated", "message": "unauthenticated", "data": null}`))
	})

	t.Run("unauthenticated response when authentication type is invalid", func(t *testing.T) {
		sessions.TokenCache.Add("a", 100, time.Minute)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Add("cookie", "sec_token=a")
		status, body, _ := testinfra.ExecuteRequest(req, engine)
		Expect(status).To(Equal(http.StatusUnauthorized))
		Expect(body).To(MatchJSON(`{"code":"security.unauthenticated", "message": "unauthenticated", "data": null}`))
	})

	t.Run("access is granted when token and authentication both valid", func(t *testing.T) {
		sessions.TokenCache.Add("b", &sessions.Session{Token: "b"}, time.Minute)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Add("cookie", "sec_token=b")
		status, body, _ := testinfra.ExecuteRequest(req, engine)
		Expect(status).To(Equal(http.StatusOK))
		Expect(body).To(Equal("b"))
	})
}

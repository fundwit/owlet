package fail_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"owlet/server/infra/fail"
	"owlet/server/testinfra"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/go-sql-driver/mysql"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

func TestPanicHandling(t *testing.T) {
	RegisterTestingT(t)

	t.Run("should be able to handle panic with error", func(t *testing.T) {
		r := gin.Default()
		r.Use(fail.ErrorHandling())

		r.GET("/", func(c *gin.Context) { panic(fmt.Errorf("some error")) })
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		status, body, _ := testinfra.ExecuteRequest(req, r)
		Expect(status).To(Equal(http.StatusInternalServerError))
		Expect(body).To(MatchJSON(`{"code":"` + fail.ErrUnexpected.Error() + `", "message":"some error", "data": null}`))
	})

	t.Run("should be able to handle panic with other object", func(t *testing.T) {
		r := gin.Default()
		r.Use(fail.ErrorHandling())

		r.GET("/", func(c *gin.Context) { panic("some error") })
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		status, body, _ := testinfra.ExecuteRequest(req, r)
		Expect(status).To(Equal(http.StatusInternalServerError))
		Expect(body).To(MatchJSON(`{"code":"` + fail.ErrUnexpected.Error() + `", "message":"some error", "data": null}`))
	})

	t.Run("should be able to handle panic with biz error", func(t *testing.T) {
		r := gin.Default()
		r.Use(fail.ErrorHandling())

		r.GET("/", func(c *gin.Context) {
			panic(&demoError{Message: "some message in demo error", Data: 1234})
		})
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		status, body, _ := testinfra.ExecuteRequest(req, r)
		Expect(status).To(Equal(444))
		Expect(body).To(MatchJSON(`{"code":"common.demo", "message":"demo error: some message in demo error", "data": 1234}`))
	})

	t.Run("should not be able to handle panic with nil", func(t *testing.T) {
		r := gin.Default()
		r.Use(fail.ErrorHandling())

		r.GET("/", func(c *gin.Context) { panic(nil) })
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		status, body, _ := testinfra.ExecuteRequest(req, r)
		Expect(status).To(Equal(http.StatusOK))
		Expect(body).To(Equal(""))
	})
}

func TestGinErrorHandling(t *testing.T) {
	RegisterTestingT(t)

	t.Run("should be able to handle error in gin.Context.Errors", func(t *testing.T) {
		r := gin.Default()
		r.Use(fail.ErrorHandling())

		r.GET("/", func(c *gin.Context) {
			c.Errors = append(c.Errors, &gin.Error{Err: errors.New("error1")})
			c.Errors = append(c.Errors, &gin.Error{Err: errors.New("error2")})
		})
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		status, body, _ := testinfra.ExecuteRequest(req, r)
		Expect(status).To(Equal(http.StatusInternalServerError))
		Expect(body).To(MatchJSON(`{"code":"` + fail.ErrUnexpected.Error() + `", "message":"error2", "data": null}`))
	})

	t.Run("should be able to handle panic error first even gin.Context.Errors is not empty", func(t *testing.T) {
		r := gin.Default()
		r.Use(fail.ErrorHandling())

		r.GET("/", func(c *gin.Context) {
			c.Errors = append(c.Errors, &gin.Error{Err: errors.New("error1")})
			panic("panic error")
		})
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		status, body, _ := testinfra.ExecuteRequest(req, r)
		Expect(status).To(Equal(http.StatusInternalServerError))
		Expect(body).To(MatchJSON(`{"code":"` + fail.ErrUnexpected.Error() + `", "message":"panic error", "data": null}`))
	})

	t.Run("should handle gin.Context.Errors when panic nil", func(t *testing.T) {
		r := gin.Default()
		r.Use(fail.ErrorHandling())

		r.GET("/", func(c *gin.Context) {
			c.Errors = append(c.Errors, &gin.Error{Err: errors.New("error1")})
			panic(nil)
		})
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		status, body, _ := testinfra.ExecuteRequest(req, r)
		Expect(status).To(Equal(http.StatusInternalServerError))
		Expect(body).To(MatchJSON(`{"code":"` + fail.ErrUnexpected.Error() + `", "message":"error1", "data": null}`))
	})
}

func TestCommonErrorHandling(t *testing.T) {
	RegisterTestingT(t)

	t.Run("should handle common.ErrForbidden", func(t *testing.T) {
		r := gin.Default()
		r.Use(fail.ErrorHandling())

		r.GET("/", func(c *gin.Context) {
			_ = c.Error(fail.ErrForbidden)
		})
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		status, body, _ := testinfra.ExecuteRequest(req, r)
		Expect(status).To(Equal(http.StatusForbidden))
		Expect(body).To(MatchJSON(`{"code":"security.forbidden", "message":"access forbidden", "data": null}`))
	})

	t.Run("should handle ErrUnauthenticated", func(t *testing.T) {
		r := gin.Default()
		r.Use(fail.ErrorHandling())

		r.GET("/", func(c *gin.Context) {
			_ = c.Error(fail.ErrUnauthenticated)
		})
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		status, body, _ := testinfra.ExecuteRequest(req, r)
		Expect(status).To(Equal(http.StatusUnauthorized))
		Expect(body).To(MatchJSON(`{"code":"security.unauthenticated", "message":"unauthenticated", "data": null}`))
	})
}

func TestThirdpartErrorHandling(t *testing.T) {
	RegisterTestingT(t)

	t.Run("should handle gorm.ErrRecordNotFound", func(t *testing.T) {
		r := gin.Default()
		r.Use(fail.ErrorHandling())

		r.GET("/", func(c *gin.Context) {
			c.Error(gorm.ErrRecordNotFound)
		})
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		status, body, _ := testinfra.ExecuteRequest(req, r)
		Expect(status).To(Equal(http.StatusNotFound))
		Expect(body).To(MatchJSON(`{"code":"common.record_not_found", "message":"record not found", "data": null}`))
	})

	t.Run("should handle io.EOF", func(t *testing.T) {
		r := gin.Default()
		r.Use(fail.ErrorHandling())

		r.GET("/", func(c *gin.Context) {
			c.Error(io.EOF)
		})
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		status, body, _ := testinfra.ExecuteRequest(req, r)
		Expect(status).To(Equal(http.StatusBadRequest))
		Expect(body).To(MatchJSON(`{"code":"bad_request.body_not_found", "message":"body not found", "data": null}`))
	})

	t.Run("should handle json.SyntaxError", func(t *testing.T) {
		r := gin.Default()
		r.Use(fail.ErrorHandling())

		r.GET("/", func(c *gin.Context) {
			c.Error(&json.SyntaxError{})
		})
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		status, body, _ := testinfra.ExecuteRequest(req, r)
		Expect(status).To(Equal(http.StatusBadRequest))
		Expect(body).To(MatchJSON(`{"code":"bad_request.invalid_body_format", "message":"invalid body format", "data": ""}`))
	})

	t.Run("should handle validator.ValidationErrors", func(t *testing.T) {
		r := gin.Default()
		r.Use(fail.ErrorHandling())

		r.GET("/", func(c *gin.Context) {
			c.Error(validator.ValidationErrors{})
		})
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		status, body, _ := testinfra.ExecuteRequest(req, r)
		Expect(status).To(Equal(http.StatusBadRequest))
		Expect(body).To(MatchJSON(`{"code":"bad_request.validation_failed", "message":"validation failed", "data": ""}`))
	})

	t.Run("should handle modify behind error", func(t *testing.T) {
		r := gin.Default()
		r.Use(fail.ErrorHandling())

		r.GET("/", func(c *gin.Context) {
			c.Error(fail.ErrModifyBehind)
		})
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		status, body, _ := testinfra.ExecuteRequest(req, r)
		Expect(status).To(Equal(http.StatusBadRequest))
		Expect(body).To(MatchJSON(`{"code":"bad_request.modify_behind", "message":"modify behind others", "data": null}`))
	})

	t.Run("should handle mysql.ErrInvalidConn", func(t *testing.T) {
		r := gin.Default()
		r.Use(fail.ErrorHandling())

		r.GET("/", func(c *gin.Context) {
			c.Error(mysql.ErrInvalidConn)
		})
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		status, body, _ := testinfra.ExecuteRequest(req, r)
		Expect(status).To(Equal(http.StatusServiceUnavailable))
		Expect(body).To(MatchJSON(`{"code":"common.internal_server_error", "message":"invalid connection", "data": null}`))
	})
}

type demoError struct {
	Message string
	Data    interface{}
}

func (e *demoError) Error() string {
	return fmt.Sprintf("demo error: %s", e.Message)
}
func (e *demoError) Respond() *fail.BizErrorDetail {
	return &fail.BizErrorDetail{
		Status: 444, Code: "common.demo",
		Message: e.Error(), Data: e.Data,
	}
}

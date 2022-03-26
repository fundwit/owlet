package misc

import (
	"net/http"
	"net/http/httptest"
	"owlet/server/testinfra"
	"testing"

	"github.com/fundwit/go-commons/types"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
)

func TestNewIdObject(t *testing.T) {
	RegisterTestingT(t)

	t.Run("NewIdObject should work as expected", func(t *testing.T) {
		Expect(*NewIdObject(100)).To(Equal(IdObject{100}))
	})
}

func TestBindingPathID(t *testing.T) {
	RegisterTestingT(t)

	t.Run("should be able to bind path id", func(t *testing.T) {
		router := gin.Default()

		var id types.ID
		var parseErr error
		router.GET("/test/:id", func(c *gin.Context) {
			id, parseErr = BindingPathID(c)
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test/123", nil)
		status, _, _ := testinfra.ExecuteRequest(req, router)
		Expect(status).To(Equal(http.StatusOK))
		Expect(id).To(Equal(types.ID(123)))
		Expect(parseErr).ToNot(HaveOccurred())
	})

	t.Run("should return err when path variable 'id' not exist", func(t *testing.T) {
		router := gin.Default()

		var id types.ID
		var parseErr error
		router.GET("/test/:xid", func(c *gin.Context) {
			id, parseErr = BindingPathID(c)
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test/123", nil)
		status, _, _ := testinfra.ExecuteRequest(req, router)
		Expect(status).To(Equal(http.StatusOK))
		Expect(id).To(BeZero())
		Expect(parseErr).To(HaveOccurred())
		Expect(parseErr.Error()).To(Equal("Key: 'requestPath.ID' Error:Field validation for 'ID' failed on the 'required' tag"))
	})

	t.Run("should return err when path variable 'id' invalid", func(t *testing.T) {
		router := gin.Default()

		var id types.ID
		var parseErr error
		router.GET("/test/:id", func(c *gin.Context) {
			id, parseErr = BindingPathID(c)
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test/abc", nil)
		status, _, _ := testinfra.ExecuteRequest(req, router)
		Expect(status).To(Equal(http.StatusOK))
		Expect(id).To(BeZero())
		Expect(parseErr).To(HaveOccurred())
		Expect(parseErr.Error()).To(Equal("invalid id 'abc'"))
	})
}

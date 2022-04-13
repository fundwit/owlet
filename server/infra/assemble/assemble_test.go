package assemble

import (
	"testing"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
)

func TestBootstrap(t *testing.T) {
	RegisterTestingT(t)

	t.Run("restful api routes should be registered as expected", func(t *testing.T) {
		engine := gin.New()
		count := 0
		for _, registerEntry := range RestAPIRegistry {
			registerEntry.Register(engine, registerEntry.MiddleWares...)
			count++
		}
		Expect(count).Should(Equal(5))
	})
}

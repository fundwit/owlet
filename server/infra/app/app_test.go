package app

import (
	"os"
	"owlet/server/infra/meta"
	"testing"

	. "github.com/onsi/gomega"
)

func TestBootstrap(t *testing.T) {
	RegisterTestingT(t)

	t.Run("bootstrap", func(t *testing.T) {

	})
}

func TestRunApp(t *testing.T) {
	RegisterTestingT(t)

	t.Run("RunApp should work as expected with default args", func(t *testing.T) {
		called := 0
		BootstrapFunc = func() {
			called++
		}

		meta.Config = nil
		os.Args = []string{"owlet"}
		RunApp()

		Expect(meta.Config.AdminName).To(Equal("admin"))
		Expect(meta.Config.AdminSecret).To(Equal("admin"))
		Expect(called).To(Equal(1))
	})

	t.Run("RunApp should work as expected", func(t *testing.T) {
		called := 0
		BootstrapFunc = func() {
			called++
		}

		meta.Config = nil
		os.Args = []string{"owlet", "--secret", "test-admin-secret", "--admin", "test-admin-name"}
		RunApp()

		Expect(meta.Config.AdminName).To(Equal("test-admin-name"))
		Expect(meta.Config.AdminSecret).To(Equal("test-admin-secret"))
		Expect(called).To(Equal(1))
	})

	t.Run("RunApp should work as expected", func(t *testing.T) {
		called := 0
		BootstrapFunc = func() {
			called++
		}

		meta.Config = nil
		os.Args = []string{"owlet", "-h"}
		RunApp()

		Expect(meta.Config).To(BeNil())
		Expect(called).To(Equal(0))
	})
}

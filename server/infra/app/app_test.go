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

		meta.Config = meta.ServiceConfig{}
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

		meta.Config = meta.ServiceConfig{}
		os.Args = []string{"owlet", "--secret", "test-admin-secret", "--admin", "test-admin-name"}
		RunApp()

		Expect(meta.Config.AdminName).To(Equal("test-admin-name"))
		Expect(meta.Config.AdminSecret).To(Equal("test-admin-secret"))
		Expect(called).To(Equal(1))
	})

	t.Run("RunApp should work as expected with env resolve", func(t *testing.T) {
		called := 0
		BootstrapFunc = func() {
			called++
		}

		os.Setenv("ENV_SECRET", "test-admin-secret1")
		os.Setenv("ENV_NAME", "test-admin-secret1")
		meta.Config = meta.ServiceConfig{}
		os.Args = []string{"owlet", "--secret", "$ENV_SECRET", "--admin", "$ENV_NAME"}
		RunApp()

		Expect(meta.Config.AdminName).To(Equal("test-admin-secret1"))
		Expect(meta.Config.AdminSecret).To(Equal("test-admin-secret1"))
		Expect(called).To(Equal(1))
	})

	t.Run("RunApp should work as expected with empty result after env resolve", func(t *testing.T) {
		called := 0
		BootstrapFunc = func() {
			called++
		}

		os.Setenv("ENV_NAME", "  ")
		os.Setenv("ENV_SECRET", "  \n")
		meta.Config = meta.ServiceConfig{}
		os.Args = []string{"owlet", "--secret", "$ENV_SECRET", "--admin", "$ENV_NAME"}
		err := RunApp()
		Expect(err).ToNot(BeNil())
		Expect(err.Error()).To(Equal("admin name is empty after env expand"))

		Expect(meta.Config.AdminName).To(Equal(""))
		Expect(meta.Config.AdminSecret).To(Equal("admin"))
		Expect(called).To(Equal(0))

		os.Setenv("ENV_NAME", " aaa ")
		meta.Config = meta.ServiceConfig{}
		os.Args = []string{"owlet", "--secret", "$ENV_SECRET", "--admin", "$ENV_NAME"}
		err = RunApp()
		Expect(err).ToNot(BeNil())
		Expect(err.Error()).To(Equal("admin secret is empty after env expand"))

		Expect(meta.Config.AdminName).To(Equal("aaa"))
		Expect(meta.Config.AdminSecret).To(Equal(""))
		Expect(called).To(Equal(0))

	})

	t.Run("RunApp should work as expected", func(t *testing.T) {
		called := 0
		BootstrapFunc = func() {
			called++
		}

		meta.Config = meta.ServiceConfig{}
		os.Args = []string{"owlet", "-h"}
		RunApp()

		Expect(meta.Config).To(Equal(meta.ServiceConfig{}))
		Expect(called).To(Equal(0))
	})
}

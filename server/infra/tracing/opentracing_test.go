package tracing_test

import (
	"owlet/server/infra/tracing"
	"testing"

	. "github.com/onsi/gomega"
	jaeger "github.com/uber/jaeger-client-go"
)

func TestNewTracer(t *testing.T) {
	RegisterTestingT(t)

	t.Run("be able to build new tracer", func(t *testing.T) {
		tr, c, err := tracing.NewTracer()
		defer c.Close()
		Expect(err).To(BeNil())
		Expect(tr).ToNot(BeNil())

		tracer, ok := tr.(*jaeger.Tracer)
		Expect(ok).To(BeTrue())
		Expect(tracer).ToNot(BeNil())
		tags := tracer.Tags()
		Expect(len(tags)).To(Equal(3)) // jaeger.version, hostname, ip

		// TODO assert effective configuration:
		//      serviceName, sampler config, reporter config, tracer config
	})
}

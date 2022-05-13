package meta

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestAcquireServiceIdentity(t *testing.T) {
	RegisterTestingT(t)

	t.Run("acquire service identity successfully", func(t *testing.T) {
		r := AcquireServiceIdentity()
		Expect(r).To(Equal(ServiceIdentity{
			Name:       "owlet",
			InstanceID: serviceInstanceId,
		}))
	})
}

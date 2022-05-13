package meta

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestAcquireServiceMeta(t *testing.T) {
	RegisterTestingT(t)

	t.Run("acquire service meta info successfully", func(t *testing.T) {
		AcquireServiceIdentityFunc = func() ServiceIdentity {
			return ServiceIdentity{Name: "owlet-test", InstanceID: "12345"}
		}
		AcquireBuildInfoFunc = func(path string) (*Build, error) {
			return &Build{Release: "test-release"}, nil
		}
		AcquireRuntimeStatusFunc = func() RuntimeStatus {
			return RuntimeStatus{NumCPU: 33}
		}
		AcquireBrandInfoFunc = func(path string) (*Brand, error) {
			return &Brand{Name: "test brand"}, nil
		}
		r := AcquireServiceMeta()
		Expect(r).To(Equal(ServiceInfo{
			ServiceIdentity: ServiceIdentity{Name: "owlet-test", InstanceID: "12345"},
			Build:           &Build{Release: "test-release"},
			Runtime:         &RuntimeStatus{NumCPU: 33},
			Brand:           &Brand{Name: "test brand"},
		}))
	})
}

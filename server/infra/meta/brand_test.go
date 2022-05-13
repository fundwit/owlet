package meta

import (
	"owlet/server/testinfra"
	"testing"

	. "github.com/onsi/gomega"
)

func TestAcquireBrandInfo(t *testing.T) {
	RegisterTestingT(t)

	t.Run("return error of file read", func(t *testing.T) {
		b, err := AcquireBrandInfo("/not-exist-file")
		Expect(b).To(BeNil())
		Expect(err.Error()).To(Equal("open /not-exist-file: no such file or directory"))
	})

	t.Run("return error of json decode", func(t *testing.T) {
		f, err := testinfra.NewFileWithContent("invalid-json-file", "invalid json content")
		Expect(err).To(BeNil())
		defer f.Clear()

		b, err := AcquireBrandInfo(f.GetRealPath())
		Expect(b).To(BeNil())
		Expect(err.Error()).To(Equal("invalid character 'i' looking for beginning of value"))
	})

	t.Run("acquire brand info successfully", func(t *testing.T) {
		f, err := testinfra.NewFileWithContent("json-file", `{
			"name": "test company",
			"logo": "test logo",
			"copyright": "test copyright",
			"license": "test license"
		  }`)
		Expect(err).To(BeNil())
		defer f.Clear()

		b, err := AcquireBrandInfo(f.GetRealPath())
		Expect(err).To(BeNil())
		Expect(*b).To(Equal(Brand{
			Name:      "test company",
			Logo:      "test logo",
			Copyright: "test copyright",
			License:   "test license",
		}))
	})
}

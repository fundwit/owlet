package meta

import (
	"owlet/server/testinfra"
	"testing"
	"time"

	"github.com/fundwit/go-commons/types"
	. "github.com/onsi/gomega"
)

func TestAcquireBuildInfo(t *testing.T) {
	RegisterTestingT(t)

	t.Run("return error of file read", func(t *testing.T) {
		b, err := AcquireBuildInfo("/not-exist-file")
		Expect(b).To(BeNil())
		Expect(err.Error()).To(Equal("open /not-exist-file: no such file or directory"))
	})

	t.Run("return error of json decode", func(t *testing.T) {
		f, err := testinfra.NewFileWithContent("invalid-json-file", "invalid json content")
		Expect(err).To(BeNil())
		defer f.Clear()

		b, err := AcquireBuildInfo(f.GetRealPath())
		Expect(b).To(BeNil())
		Expect(err.Error()).To(Equal("invalid character 'i' looking for beginning of value"))
	})

	t.Run("acquire build info successfully", func(t *testing.T) {
		f, err := testinfra.NewFileWithContent("json-file", `{
			"buildTime": "2022-04-25T8:34:54Z",
			"version": "stage.4bc144742c",
			"sourceCodes": [{
			  "repository": "test/test-repo",
			  "ref": "refs/heads/stage",
			  "reversion": {
				"id": "4bc144742cb883e01196eb157b280873145e19d0",
				"author": "test author",
				"title": "test title",
				"message": "test message",
				"timestamp": "2022-04-20T12:30:40+08:00"
			  }
			}]
		  }`)
		Expect(err).To(BeNil())
		defer f.Clear()

		b, err := AcquireBuildInfo(f.GetRealPath())
		b.Timestamp = types.Timestamp(b.Timestamp.Time().In(time.UTC))
		b.SourceCodes[0].LastChange.Timestamp = types.Timestamp(b.SourceCodes[0].LastChange.Timestamp.Time().In(time.UTC))
		Expect(err).To(BeNil())
		Expect(*b).To(Equal(Build{
			Timestamp: types.TimestampOfDate(2022, 4, 25, 8, 34, 54, 0, time.UTC),
			Release:   "stage.4bc144742c",
			SourceCodes: []SourceCode{
				{
					Repository: "test/test-repo",
					Ref:        "refs/heads/stage",
					LastChange: CodeReversion{
						ID:        "4bc144742cb883e01196eb157b280873145e19d0",
						Timestamp: types.TimestampOfDate(2022, 4, 20, 4, 30, 40, 0, time.UTC),
						Author:    "test author",
						Title:     "test title",
						Message:   "test message",
					},
				},
			},
		}))
	})
}

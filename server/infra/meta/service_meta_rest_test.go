package meta

import (
	"net/http"
	"net/http/httptest"
	"owlet/server/infra/fail"
	"owlet/server/testinfra"
	"testing"
	"time"

	"github.com/fundwit/go-commons/types"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
)

func TestQueryTagsAPI(t *testing.T) {
	RegisterTestingT(t)

	router := gin.Default()
	router.Use(fail.ErrorHandling())
	RegisterMetaRestAPI(router)

	t.Run("should be able to handle query service meta info successfully", func(t *testing.T) {
		AcquireServiceMetaFunc = func() ServiceInfo {
			return ServiceInfo{
				ServiceIdentity: ServiceIdentity{Name: "owlet-test", InstanceID: "12345"},
				Build: &Build{
					Release:   "test release",
					Timestamp: types.TimestampOfDate(2022, 1, 2, 11, 30, 40, 0, time.UTC),
					SourceCodes: []SourceCode{
						{
							Repository: "http://test-repo.com/test-repo.git",
							Ref:        "master",
							LastChange: CodeReversion{
								ID:        "abcde",
								Timestamp: types.TimestampOfDate(2022, 1, 2, 10, 30, 30, 0, time.UTC),
								Author:    "ann",
								Title:     "Fix xxx",
								Message:   "xxxxxxx",
							},
						},
					},
				},
				Runtime: &RuntimeStatus{
					StartTime:    types.TimestampOfDate(2022, 1, 2, 12, 0, 0, 0, time.UTC),
					NumCPU:       33,
					NumGoroutine: 111,
					NumMaxProcs:  222,
				},
				Brand: &Brand{
					Name:      "test-company",
					Logo:      "test logo",
					Copyright: "test copyright",
					License:   "test license",
				},
			}
		}
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		status, body, _ := testinfra.ExecuteRequest(req, router)
		Expect(status).To(Equal(http.StatusOK))
		Expect(body).To(MatchJSON(`{
			"name": "owlet-test",
			"instanceId": "12345",
			"build": {
			  "buildTime": "2022-01-02T11:30:40Z",
			  "version": "test release",
			  "sourceCodes": [
				{
				  "repository": "http://test-repo.com/test-repo.git",
				  "ref": "master",
				  "reversion": {
					"id": "abcde",
					"timestamp": "2022-01-02T10:30:30Z",
					"author": "ann",
					"title": "Fix xxx",
					"message": "xxxxxxx"
				  }
				}
			  ]
			},
			"runtime": {
			  "startTime": "2022-01-02T12:00:00Z",
			  "numCpu": 33,
			  "numGoroutine": 111,
			  "numMaxProcs": 222
			},
			"brand": {
			  "name": "test-company",
			  "logo": "test logo",
			  "copyright": "test copyright",
			  "license": "test license"
			}
		  }`))
	})
}

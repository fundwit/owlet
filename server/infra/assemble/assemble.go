package assemble

import (
	"owlet/server/domain"
	"owlet/server/infra/doc"
	"owlet/server/infra/meta"
	"owlet/server/infra/sessions"

	"github.com/gin-gonic/gin"
)

/*
* registry endpoint for:
*
*   1. database auto migrations
*   2. rest api routes
*   3. error serialize
*   4. metric collectors
 */

type RestAPIRegister func(*gin.Engine, ...gin.HandlerFunc)

var AutoMigrations = []interface{}{}
var RestAPIRegistry = []APIRegistryEntry{}

type APIRegistryEntry struct {
	Register    RestAPIRegister
	MiddleWares []gin.HandlerFunc
}

func init() {
	AutoMigrations = []interface{}{}
	RestAPIRegistry = []APIRegistryEntry{
		{meta.RegisterMetaRestAPI, nil},
		{doc.RegisterDocsAPI, nil},
		{domain.RegisterArticlesRestAPI, []gin.HandlerFunc{sessions.SessionTokenAuth()}},
		{domain.RegisterTagsRestAPI, nil},
		{domain.RegisterTagAssignsRestAPI, []gin.HandlerFunc{sessions.SessionTokenAuth()}},
	}
}

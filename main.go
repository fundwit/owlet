package main

import (
	"owlet/server/infra/app"
)

// @Title owlet
// @version v0.1.x
// @Description A Wiki services.
// @Accept  json
// @Produce  json
func main() {
	if err := app.RunAppFunc(); err != nil {
		panic(err)
	}
}

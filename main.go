package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// serve index page
	r.StaticFile("", "./ui/index.html")

	// serve javascript webgl
	r.StaticFile("/js/webgl.js", "./src/js/webgl.js")

	r.Run("localhost:8080")
}

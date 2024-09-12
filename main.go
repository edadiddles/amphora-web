package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// index page
	r.StaticFile("", "./ui/index.html")

	r.Run("localhost:8080")
}

package main

import (
	"math/rand/v2"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// serve index page
	r.StaticFile("", "./ui/index.html")

	// serve javascript webgl
	r.StaticFile("/js/webgl.js", "./src/js/webgl.js"))

	// api
	r.GET("/api/verticies", GetVerticies)
	r.Run("localhost:8080")
}

func GetVerticies(c *gin.Context) {
	numVerticies, err := strconv.Atoi(c.Query("numVerticies"))
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	verticies := generateVerticies(numVerticies)
	c.JSON(http.StatusOK, gin.H{
		"numVerticies": numVerticies,
		"verticies":    verticies,
	})
}

func generateVerticies(numVerticies int) []float32 {
	verticies := make([]float32, 3*numVerticies)

	for i := 0; i < numVerticies; i++ {
		verticies[3*i] = rand.Float32()*10 - 5
		verticies[3*i+1] = rand.Float32()*10 - 5
		verticies[3*i+2] = rand.Float32()*10 - 5
	}
	return verticies
}

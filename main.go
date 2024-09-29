package main

import (
	"encoding/xml"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/gin-contrib/pprof"
)

type Phone struct {
	Width   float32 `xml:"width"`
	Length  float32 `xml:"length"`
	Height  float32 `xml:"height"`
	Speaker Speaker `xml:"Speaker"`
}

type Speaker struct {
	Width  float32 `xml:"width"`
	Height float32 `xml:"height"`
	Center float32 `xml:"center"`
}

func parseXml(xmlFile string) (*Phone, error) {
	f, err := os.Open(xmlFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	byteValue, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var phone Phone
	xml.Unmarshal(byteValue, &phone)

	return &phone, err
}

func main() {
	r := gin.Default()

	// serve index page
	r.StaticFile("", "./ui/index.html")

	// serve javascript webgl
	r.StaticFile("/js/webgl.js", "./src/js/webgl.js")

	// api
	r.GET("/api/verticies", GetVerticies)
	r.GET("/api/phone", GetPhoneDimensions)

	// profiler registrations
	pprof.Register(r)

	// run web server
	r.Run("localhost:8080")
}

func GetPhoneDimensions(c *gin.Context) {
	model := c.Query("model")

	var filepath = model + ".xml"

	phone, err := parseXml(filepath)
	if err != nil {
		c.AbortWithStatus(http.StatusTeapot)
		return
	}

	c.JSON(http.StatusOK, phone)
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

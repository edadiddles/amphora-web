package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/gin-contrib/pprof"
)

type PhoneOption struct {
	Name     string
	Filename string
}

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

type Paraboloid struct {
	X     float32 `json:"x"`
	Y     float32 `json:"y"`
	Z     float32 `json:"z"`
	Angle float32 `json:"angle"`
}

type SlicingPlane struct {
	Height float32 `json:"height"`
	Angle  float32 `json:"angle"`
}

type UserRadius struct {
	Radius float32 `json:"user"`
}

type Resolution struct {
	Linear  float32 `json:"linear"`
	Angular float32 `json:"angular"`
}

type SimulationInput struct {
	Phone        string       `json:"phone"`
	Paraboloid   Paraboloid   `json:"paraboloid"`
	SlicingPlane SlicingPlane `json:"slicingPlane"`
	UserRadius   UserRadius   `json:"userRadius"`
	Resolution   Resolution   `json:"resolution"`
}

type SimulationOutput struct {
	Phone      []float32
	Paraboloid []float32
	User       []float32
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

	// htmx
	r.GET("/htmx/phones", HandleHtmxGetPhones)

	// api
	r.GET("/api/phone", HandleApiPhone)
	r.GET("/api/phones", HandleApiGetPhones)

	r.POST("/api/simulation", HandleApiSimulation)

	// profiler registrations
	pprof.Register(r)

	// run web server
	r.Run("localhost:8080")
}

func HandleApiSimulation(c *gin.Context) {
	var simulationInput SimulationInput
	err := c.BindJSON(&simulationInput)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	phone, err := getPhoneDimensions(simulationInput.Phone)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	phoneVerticies, paraboloidVerticies, userVerticies, err := generateSimulation(phone, &simulationInput.Paraboloid, &simulationInput.SlicingPlane, &simulationInput.UserRadius, &simulationInput.Resolution)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	var simulationOutput SimulationOutput
	simulationOutput.Phone = phoneVerticies
	simulationOutput.Paraboloid = paraboloidVerticies
	simulationOutput.User = userVerticies

	c.JSON(http.StatusOK, simulationOutput)
}

func generateSimulation(phone *Phone, paraboloid *Paraboloid, slicingPlane *SlicingPlane, userRadius *UserRadius, resolution *Resolution) ([]float32, []float32, []float32, error) {
	// setup scene
	// phone bottom center should intersect the 'normal' vector of the paraboloids 'focus plane'
	// the paraboloids focus plane is the plane that includes the paraboloids focus and normal vector intersects the vertex
	// the phone should have all four corners touch the surface of the parabolid
	// the slicing plane represents the ending of the paraboloid surface
	// the user radius indicates the distance from which the user will be positioned from the amphora device
	// the sound will be simulated as particles of sounds (aka phonons) and interference effects will be ignored
	// the sound will be emitted along a rectangular speaker that is split into a grid and will assume the speaker will emit sound at a specified angle
	// the speaker grid will be split into equal rows and columns based on the 'linear resolution'
	// the speaker emission cone will be split into equal azimuth and altitude sections based on the 'angular resolution'
	// there there will be (linear resolution)^2 points of sound emission and (angular resolution)^2 directional emissions of sound
	// hence there will be (linear resolution)^2*(angular resolution)^2 sound particles tracked in this simulation.

	phoneVerticies := generateVerticies(10000)
	paraboloidVerticies := generateVerticies(0)
	userVerticies := generateVerticies(0)

	return phoneVerticies, paraboloidVerticies, userVerticies, nil
}

func HandleHtmxGetPhones(c *gin.Context) {
	phoneOptions, err := GetPhones()
	if err != nil {
		c.String(http.StatusInternalServerError, "")
	}

	htmlOptions := ""
	for i := 0; i < len(phoneOptions); i++ {
		htmlOptions += fmt.Sprintf("<option id='%s' value='%s'>%s</option>", phoneOptions[i].Name, phoneOptions[i].Filename, phoneOptions[i].Name)
	}

	c.String(http.StatusOK, htmlOptions)
}

func HandleApiPhone(c *gin.Context) {
	model := c.Query("model")
	var filepath = model + ".xml"

	phone, err := getPhoneDimensions(filepath)
	if err != nil {
		c.AbortWithStatus(http.StatusTeapot)
		return
	}

	c.JSON(http.StatusOK, phone)
}

func HandleApiGetPhones(c *gin.Context) {
	phoneOptions, err := GetPhones()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	c.JSON(http.StatusOK, phoneOptions)
}

func GetPhones() ([]PhoneOption, error) {
	dir, err := os.ReadDir("phones")
	if err != nil {
		return nil, err
	}

	var phoneOptions []PhoneOption
	for i := 0; i < len(dir); i++ {
		if dir[i].IsDir() {
			// WHY?
			continue
		}

		if strings.Contains(dir[i].Name(), ".xml") {
			var option PhoneOption
			option.Filename = dir[i].Name()
			option.Name = strings.Replace(dir[i].Name(), ".xml", "", -1)
			phoneOptions = append(phoneOptions, option)
		}
	}
	return phoneOptions, nil
}

func getPhoneDimensions(filepath string) (*Phone, error) {
	phone, err := parseXml("phones/" + filepath)
	if err != nil {
		return nil, err
	}

	return phone, nil

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
		x := rand.Float32()*10 - 5
		y := rand.Float32()*2 - 1
		z := rand.Float32()*2 - 1

		verticies[3*i] = x
		verticies[3*i+1] = y
		verticies[3*i+2] = z //x*x + y*y
	}
	return verticies
}

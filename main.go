package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
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

type PhoneConfig struct {
	Width   float64       `xml:"width"`
	Length  float64       `xml:"length"`
	Height  float64       `xml:"height"`
	Speaker SpeakerConfig `xml:"Speaker"`
}

type SpeakerConfig struct {
	Width  float64 `xml:"width"`
	Height float64 `xml:"height"`
	Center float64 `xml:"center"`
}

type PhoneInput struct {
	Filename string  `json:"filename"`
	Angle    float64 `json:"angle"`
}

type ParaboloidInput struct {
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Z     float64 `json:"z"`
	Angle float64 `json:"angle"`
}

type SlicingPlaneInput struct {
	Height float64 `json:"height"`
	Angle  float64 `json:"angle"`
}

type UserRadiusInput struct {
	Radius float64 `json:"user"`
}

type ResolutionInput struct {
	Linear  float64 `json:"linear"`
	Angular float64 `json:"angular"`
}

type SimulationInput struct {
	Phone        PhoneInput        `json:"phone"`
	Paraboloid   ParaboloidInput   `json:"paraboloid"`
	SlicingPlane SlicingPlaneInput `json:"slicingPlane"`
	UserRadius   UserRadiusInput   `json:"userRadius"`
	Resolution   ResolutionInput   `json:"resolution"`
}

type SimulationOutput struct {
	Phone      []float64
	Paraboloid []float64
	User       []float64
}

func parseXml(xmlFile string) (*PhoneConfig, error) {
	f, err := os.Open(xmlFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	byteValue, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var phone PhoneConfig
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

	phoneConfig, err := getPhoneDimensions(simulationInput.Phone.Filename)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	phoneVerticies, paraboloidVerticies, userVerticies, err := generateSimulation(phoneConfig, &simulationInput.Phone, &simulationInput.Paraboloid, &simulationInput.SlicingPlane, &simulationInput.UserRadius, &simulationInput.Resolution)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	var simulationOutput SimulationOutput
	simulationOutput.Phone = phoneVerticies
	simulationOutput.Paraboloid = paraboloidVerticies
	simulationOutput.User = userVerticies

	c.JSON(http.StatusOK, simulationOutput)
}

func generateSimulation(phoneConfig *PhoneConfig, phoneInput *PhoneInput, paraboloidInput *ParaboloidInput, slicingPlaneInput *SlicingPlaneInput, userRadiusInput *UserRadiusInput, resolutionInput *ResolutionInput) ([]float64, []float64, []float64, error) {
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

	spanAzimuthal := math.Pi / 6
	spanPolar := 2 * math.Pi

	widthPhone := phoneConfig.Width
	lengthPhone := phoneConfig.Length
	heightPhone := phoneConfig.Height
	anglePhone := phoneInput.Angle

	speakerWidth := phoneConfig.Speaker.Width
	speakerHeight := phoneConfig.Speaker.Height
	speakerCenter := phoneConfig.Speaker.Center

	coefficientsParaboloidX := paraboloidInput.X
	coefficientsParaboloidY := paraboloidInput.Y
	coefficientsParaboloidZ := paraboloidInput.Z
	angleParaboloid := paraboloidInput.Angle

	phoneCorner := make([]float64, 3)
	phoneCorner[0] = widthPhone / 2
	phoneCorner[1] = +(-8.0*math.Pow(widthPhone, 2)*coefficientsParaboloidX*coefficientsParaboloidY*math.Sin(angleParaboloid) - 6.0*math.Pow(lengthPhone, 2)*math.Pow(coefficientsParaboloidY, 2)*math.Sin(angleParaboloid) - 8.0*math.Pow(coefficientsParaboloidZ, 2)*math.Sin(angleParaboloid) - 16.0*lengthPhone*coefficientsParaboloidY*coefficientsParaboloidZ*math.Sin(anglePhone) - 8.0*lengthPhone*coefficientsParaboloidY*coefficientsParaboloidZ*math.Sin(2.0*angleParaboloid+anglePhone) - 4.0*math.Pow(widthPhone, 2)*coefficientsParaboloidX*coefficientsParaboloidY*math.Sin(angleParaboloid+2.0*anglePhone) - 4.0*math.Pow(lengthPhone, 2)*math.Pow(coefficientsParaboloidY, 2)*math.Sin(angleParaboloid+2.0*anglePhone) + 12.0*math.Pow(coefficientsParaboloidZ, 2)*math.Sin(angleParaboloid+2.0*anglePhone) + 4.0*math.Pow(widthPhone, 2)*coefficientsParaboloidX*coefficientsParaboloidY*math.Sin(3.0*angleParaboloid+2.0*anglePhone) + 4.0*math.Pow(lengthPhone, 2)*math.Pow(coefficientsParaboloidY, 2)*math.Sin(3.0*angleParaboloid+2.0*anglePhone) + 4.0*math.Pow(coefficientsParaboloidZ, 2)*math.Sin(3.0*angleParaboloid+2.0*anglePhone) + 8.0*lengthPhone*coefficientsParaboloidY*coefficientsParaboloidZ*math.Sin(2.0*angleParaboloid+3.0*anglePhone) + math.Pow(lengthPhone, 2)*math.Pow(coefficientsParaboloidY, 2)*math.Sin(3.0*angleParaboloid+4.0*anglePhone) - math.Pow(lengthPhone, 2)*math.Pow(coefficientsParaboloidY, 2)*math.Sin(5.0*angleParaboloid+4.0*anglePhone)) / (64.0 * coefficientsParaboloidY * coefficientsParaboloidZ * math.Pow(math.Sin(angleParaboloid+anglePhone), 2))
	phoneCorner[2] = -(-8.0*math.Pow(widthPhone, 2)*math.Cos(angleParaboloid)*coefficientsParaboloidX*coefficientsParaboloidY + 4.0*math.Pow(widthPhone, 2)*math.Cos(angleParaboloid+2.0*anglePhone)*coefficientsParaboloidX*coefficientsParaboloidY + 4.0*math.Pow(widthPhone, 2)*math.Cos(3.0*angleParaboloid+2.0*anglePhone)*coefficientsParaboloidX*coefficientsParaboloidY - 6.0*math.Pow(lengthPhone, 2)*math.Cos(angleParaboloid)*math.Pow(coefficientsParaboloidY, 2) + 4.0*math.Pow(lengthPhone, 2)*math.Cos(angleParaboloid+2.0*anglePhone)*math.Pow(coefficientsParaboloidY, 2) + 4.0*math.Pow(lengthPhone, 2)*math.Cos(3.0*angleParaboloid+2.0*anglePhone)*math.Pow(coefficientsParaboloidY, 2) - math.Pow(lengthPhone, 2)*math.Cos(3.0*angleParaboloid+4.0*anglePhone)*math.Pow(coefficientsParaboloidY, 2) - math.Pow(lengthPhone, 2)*math.Cos(5.0*angleParaboloid+4.0*anglePhone)*math.Pow(coefficientsParaboloidY, 2) + 16.0*lengthPhone*math.Cos(anglePhone)*coefficientsParaboloidY*coefficientsParaboloidZ - 8.0*lengthPhone*math.Cos(2.0*angleParaboloid+anglePhone)*coefficientsParaboloidY*coefficientsParaboloidZ - 8.0*lengthPhone*math.Cos(2.0*angleParaboloid+3.0*anglePhone)*coefficientsParaboloidY*coefficientsParaboloidZ - 8.0*math.Cos(angleParaboloid)*math.Pow(coefficientsParaboloidZ, 2) - 12.0*math.Cos(angleParaboloid+2.0*anglePhone)*math.Pow(coefficientsParaboloidZ, 2) + 4.0*math.Cos(3.0*angleParaboloid+2.0*anglePhone)*math.Pow(coefficientsParaboloidZ, 2)) / (64.0 * coefficientsParaboloidY * coefficientsParaboloidZ * math.Pow(math.Sin(angleParaboloid+anglePhone), 2))

	vecPhoneWidth := []float64{1, 0, 0}

	speakerNormal := []float64{0, math.Sin(anglePhone), math.Cos(anglePhone)}
	phoneNormal := []float64{0, math.Cos(float64(anglePhone)), math.Sin(float64(anglePhone))}

	//initialize phonon
	initialPhononLoc := make([]float64, 3)
	initialPhononProj := make([]float64, 3)
	for i := 0; i < len(initialPhononLoc); i++ {
		initialPhononLoc[i] = phoneCorner[i] - speakerCenter*vecPhoneWidth[i] + 0.5*heightPhone*phoneNormal[i]
		initialPhononProj[i] = -speakerNormal[i]
	}

	linearResolution := 1 / resolutionInput.Linear
	angularResolution := 1 / resolutionInput.Angular
	var sum int64
	for gridSpeakerWidth := -0.5 * speakerWidth; gridSpeakerWidth <= 0.5*speakerWidth; gridSpeakerWidth += linearResolution {
		for gridSpeakerHeight := -0.5 * speakerHeight; gridSpeakerHeight <= 0.5*speakerHeight; gridSpeakerHeight += linearResolution {
			speakerLoc := make([]float64, 3)
			for i := 0; i < len(speakerLoc); i++ {
				speakerLoc[i] = initialPhononLoc[i] + gridSpeakerWidth*vecPhoneWidth[i] + gridSpeakerHeight*phoneNormal[i]
			}

			for gridAzimuthal := 0.0; gridAzimuthal <= spanAzimuthal; gridAzimuthal += angularResolution {
				for gridPolar := 0.0; gridPolar <= spanPolar; gridPolar += angularResolution {
					sum += 1
				}
			}
		}
	}

	fmt.Println(sum)
	fmt.Println("generating verticies...")
	phoneVerticies := generateVerticies(sum, "phone")
	paraboloidVerticies := generateVerticies(sum, "paraboloid")
	userVerticies := generateVerticies(sum, "userSphere")

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

func getPhoneDimensions(filepath string) (*PhoneConfig, error) {
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

	verticies := generateVerticies(int64(numVerticies), "")
	c.JSON(http.StatusOK, gin.H{
		"numVerticies": numVerticies,
		"verticies":    verticies,
	})
}

func generateVerticies(numVerticies int64, shape string) []float64 {
	verticies := make([]float64, 3*numVerticies)

	phi := float64(math.Pi / 6)
	theta := float64(math.Pi / 3)
	radius := float64(2.0)
	for i := int64(0); i < numVerticies; i++ {
		x := rand.Float64()*radius - radius/2
		y := rand.Float64()*radius - radius/2
		// z := rand.Float64()*2 - 1

		if shape == "paraboloid" {
			verticies[3*i] = x
			verticies[3*i+1] = y*math.Cos(phi) - (x*x+y*y)*math.Sin(phi)
			verticies[3*i+2] = y*math.Sin(phi) + (x*x+y*y)*math.Cos(phi)
		} else if shape == "phone" {
			verticies[3*i] = x
			verticies[3*i+1] = y
			verticies[3*i+2] = y * math.Tan(theta) //x*x + y*y
		} else if shape == "userSphere" {
			verticies[3*i] = x
			verticies[3*i+1] = y
			verticies[3*i+2] = math.Sqrt(radius*radius - (x*x + y*y))
		}
	}
	return verticies
}

func dotProduct(a []float64, b []float64) (float64, error) {
	if len(a) != len(b) {
		return 0.0, errors.New("inputs not same length")
	}

	sum := 0.0
	for i := 0; i < len(a); i++ {
		sum += a[i] * b[i]
	}

	return sum, nil
}

func crossProduct(a []float64, b []float64) ([]float64, error) {
	if len(a) != len(b) {
		return nil, errors.New("inputs not same length")
	} else if len(a) != 3 {
		return nil, errors.New("inputs not length 3")
	}
	product := []float64{a[1]*b[2] - a[2]*b[1], a[2]*b[0] - a[0]*b[2], a[0]*b[1] - a[1]*b[0]}
	return product, nil
}

func rotation(rotationMatrix [][]float64, axisOfRotation []float64, rotationAngle float64) {

	rotationMatrix[0][0] = math.Cos(rotationAngle) + axisOfRotation[0]*axisOfRotation[0]*(1-math.Cos(rotationAngle))
	rotationMatrix[1][1] = math.Cos(rotationAngle) + axisOfRotation[1]*axisOfRotation[1]*(1-math.Cos(rotationAngle))
	rotationMatrix[2][2] = math.Cos(rotationAngle) + axisOfRotation[2]*axisOfRotation[2]*(1-math.Cos(rotationAngle))

	rotationMatrix[0][1] = -axisOfRotation[2]*math.Sin(rotationAngle) + axisOfRotation[0]*axisOfRotation[1]*(1-math.Cos(rotationAngle))
	rotationMatrix[0][2] = axisOfRotation[1]*math.Sin(rotationAngle) + axisOfRotation[0]*axisOfRotation[2]*(1-math.Cos(rotationAngle))

	rotationMatrix[1][0] = axisOfRotation[2]*math.Sin(rotationAngle) + axisOfRotation[0]*axisOfRotation[1]*(1-math.Cos(rotationAngle))
	rotationMatrix[1][2] = -axisOfRotation[0]*math.Sin(rotationAngle) + axisOfRotation[1]*axisOfRotation[2]*(1-math.Cos(rotationAngle))

	rotationMatrix[2][0] = -axisOfRotation[1]*math.Sin(rotationAngle) + axisOfRotation[0]*axisOfRotation[2]*(1-math.Cos(rotationAngle))
	rotationMatrix[2][1] = axisOfRotation[0]*math.Sin(rotationAngle) + axisOfRotation[1]*axisOfRotation[2]*(1-math.Cos(rotationAngle))
}

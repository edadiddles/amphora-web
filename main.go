package main

import (
	"amphora/linalg"
	"encoding/xml"
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
	Radius float64 `json:"radius"`
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
	var coefficientsParaboloidX float64
	var coefficientsParaboloidY float64
	var coefficientsParaboloidZ float64
	var angleParaboloid float64
	var heightSlicingPlane float64
	var angleSlicingPlane float64

	var anglePhone float64
	var radiusUser float64
	var linearResolution float64
	var angularResolution float64

	var widthPhone float64
	var lengthPhone float64
	var heightPhone float64
	var widthSpeaker float64
	var heightSpeaker float64
	var centerSpeaker float64

	var xPhone float64
	var yPhone float64
	var zPhone float64
	cornerPhone := []float64{0, 0, 0}

	vecPhoneWidth := []float64{0, 0, 0}
	vecPhoneLength := []float64{0, 0, 0}
	vecPhoneHeight := []float64{0, 0, 0}

	initialPhononLocation := []float64{0, 0, 0}
	initialPhononProjection := []float64{0, 0, 0}

	locationSpeaker := []float64{0, 0, 0}

	spanAzimuthal := 30 * math.Pi / 180
	spanPolar := 2 * math.Pi

	locationPhonon := []float64{0, 0, 0}
	projectionPhonon := []float64{0, 0, 0}
	rotationPolar := [][]float64{
		{0, 0, 0},
		{0, 0, 0},
		{0, 0, 0},
	}
	rotationAzimuthal := [][]float64{
		{0, 0, 0},
		{0, 0, 0},
		{0, 0, 0},
	}

	atUser := false

	intersectParaboloid := []float64{0, 0, 0}
	intersectPhone := []float64{0, 0, 0}
	intersectUser := []float64{0, 0, 0}

	var aParaboloid float64
	var bParaboloid float64
	var cParaboloid float64
	var tParaboloid float64
	var tSlicingPlane float64
	var tPhone float64
	var aUser float64
	var bUser float64
	var cUser float64
	var tUser float64

	phoneVerticies := make([]float64, 0, 50000)
	paraboloidVerticies := make([]float64, 0, 50000)
	userVerticies := make([]float64, 0, 50000)

	coefficientsParaboloidX = paraboloidInput.X
	coefficientsParaboloidY = paraboloidInput.Y
	coefficientsParaboloidZ = paraboloidInput.Z
	angleParaboloid = paraboloidInput.Angle * math.Pi / 180

	heightSlicingPlane = slicingPlaneInput.Height * 10
	angleSlicingPlane = slicingPlaneInput.Angle * math.Pi / 180

	anglePhone = phoneInput.Angle * math.Pi / 180

	radiusUser = userRadiusInput.Radius * 1000

	linearResolution = resolutionInput.Linear
	angularResolution = resolutionInput.Angular

	widthPhone = phoneConfig.Width
	lengthPhone = phoneConfig.Length
	heightPhone = phoneConfig.Height
	centerSpeaker = phoneConfig.Speaker.Center
	heightSpeaker = phoneConfig.Speaker.Height
	widthSpeaker = phoneConfig.Speaker.Width

	xPhone = 0.5 * widthPhone
	yPhone = +(-8.0*math.Pow(widthPhone, 2)*coefficientsParaboloidX*coefficientsParaboloidY*math.Sin(angleParaboloid) - 6.0*math.Pow(lengthPhone, 2)*math.Pow(coefficientsParaboloidY, 2)*math.Sin(angleParaboloid) - 8.0*math.Pow(coefficientsParaboloidZ, 2)*math.Sin(angleParaboloid) - 16.0*lengthPhone*coefficientsParaboloidY*coefficientsParaboloidZ*math.Sin(anglePhone) - 8.0*lengthPhone*coefficientsParaboloidY*coefficientsParaboloidZ*math.Sin(2.0*angleParaboloid+anglePhone) - 4.0*math.Pow(widthPhone, 2)*coefficientsParaboloidX*coefficientsParaboloidY*math.Sin(angleParaboloid+2.0*anglePhone) - 4.0*math.Pow(lengthPhone, 2)*math.Pow(coefficientsParaboloidY, 2)*math.Sin(angleParaboloid+2.0*anglePhone) + 12.0*math.Pow(coefficientsParaboloidZ, 2)*math.Sin(angleParaboloid+2.0*anglePhone) + 4.0*math.Pow(widthPhone, 2)*coefficientsParaboloidX*coefficientsParaboloidY*math.Sin(3.0*angleParaboloid+2.0*anglePhone) + 4.0*math.Pow(lengthPhone, 2)*math.Pow(coefficientsParaboloidY, 2)*math.Sin(3.0*angleParaboloid+2.0*anglePhone) + 4.0*math.Pow(coefficientsParaboloidZ, 2)*math.Sin(3.0*angleParaboloid+2.0*anglePhone) + 8.0*lengthPhone*coefficientsParaboloidY*coefficientsParaboloidZ*math.Sin(2.0*angleParaboloid+3.0*anglePhone) + math.Pow(lengthPhone, 2)*math.Pow(coefficientsParaboloidY, 2)*math.Sin(3.0*angleParaboloid+4.0*anglePhone) - math.Pow(lengthPhone, 2)*math.Pow(coefficientsParaboloidY, 2)*math.Sin(5.0*angleParaboloid+4.0*anglePhone)) / (64.0 * coefficientsParaboloidY * coefficientsParaboloidZ * math.Pow(math.Sin(angleParaboloid+anglePhone), 2))
	zPhone = -(-8.0*math.Pow(widthPhone, 2)*math.Cos(angleParaboloid)*coefficientsParaboloidX*coefficientsParaboloidY + 4.0*math.Pow(widthPhone, 2)*math.Cos(angleParaboloid+2.0*anglePhone)*coefficientsParaboloidX*coefficientsParaboloidY + 4.0*math.Pow(widthPhone, 2)*math.Cos(3.0*angleParaboloid+2.0*anglePhone)*coefficientsParaboloidX*coefficientsParaboloidY - 6.0*math.Pow(lengthPhone, 2)*math.Cos(angleParaboloid)*math.Pow(coefficientsParaboloidY, 2) + 4.0*math.Pow(lengthPhone, 2)*math.Cos(angleParaboloid+2.0*anglePhone)*math.Pow(coefficientsParaboloidY, 2) + 4.0*math.Pow(lengthPhone, 2)*math.Cos(3.0*angleParaboloid+2.0*anglePhone)*math.Pow(coefficientsParaboloidY, 2) - math.Pow(lengthPhone, 2)*math.Cos(3.0*angleParaboloid+4.0*anglePhone)*math.Pow(coefficientsParaboloidY, 2) - math.Pow(lengthPhone, 2)*math.Cos(5.0*angleParaboloid+4.0*anglePhone)*math.Pow(coefficientsParaboloidY, 2) + 16.0*lengthPhone*math.Cos(anglePhone)*coefficientsParaboloidY*coefficientsParaboloidZ - 8.0*lengthPhone*math.Cos(2.0*angleParaboloid+anglePhone)*coefficientsParaboloidY*coefficientsParaboloidZ - 8.0*lengthPhone*math.Cos(2.0*angleParaboloid+3.0*anglePhone)*coefficientsParaboloidY*coefficientsParaboloidZ - 8.0*math.Cos(angleParaboloid)*math.Pow(coefficientsParaboloidZ, 2) - 12.0*math.Cos(angleParaboloid+2.0*anglePhone)*math.Pow(coefficientsParaboloidZ, 2) + 4.0*math.Cos(3.0*angleParaboloid+2.0*anglePhone)*math.Pow(coefficientsParaboloidZ, 2)) / (64.0 * coefficientsParaboloidY * coefficientsParaboloidZ * math.Pow(math.Sin(angleParaboloid+anglePhone), 2))
	cornerPhone[0] = xPhone
	cornerPhone[1] = yPhone
	cornerPhone[2] = zPhone

	vecPhoneWidth[0] = widthPhone
	vecPhoneWidth[1] = 0
	vecPhoneWidth[2] = 0
	linalg.Normalize(vecPhoneWidth, 3)

	vecPhoneLength[0] = 0
	vecPhoneLength[1] = lengthPhone * math.Sin(anglePhone)
	vecPhoneLength[2] = lengthPhone * math.Cos(anglePhone)
	linalg.Normalize(vecPhoneLength, 3)

	vecPhoneHeight[0] = vecPhoneLength[1]*vecPhoneWidth[2] - vecPhoneLength[2]*vecPhoneWidth[1]
	vecPhoneHeight[1] = vecPhoneLength[2]*vecPhoneWidth[0] - vecPhoneLength[0]*vecPhoneWidth[2]
	vecPhoneHeight[2] = vecPhoneLength[0]*vecPhoneWidth[1] - vecPhoneLength[1]*vecPhoneWidth[0]
	linalg.Normalize(vecPhoneHeight, 3)

	for i := 0; i < 3; i++ {
		initialPhononLocation[i] = cornerPhone[i] - centerSpeaker*vecPhoneWidth[i] + 0.5*heightPhone*vecPhoneHeight[i]
		initialPhononProjection[i] = -vecPhoneLength[i]
	}

	for gridSpeakerWidth := -0.5 * widthSpeaker; gridSpeakerWidth <= 0.5*widthSpeaker; gridSpeakerWidth += linearResolution {
		for gridSpeakerHeight := -0.5 * heightSpeaker; gridSpeakerHeight <= 0.5*heightSpeaker; gridSpeakerHeight += linearResolution {
			for i := 0; i < 3; i++ {
				locationSpeaker[i] = initialPhononLocation[i] + gridSpeakerWidth*vecPhoneWidth[i] + gridSpeakerHeight*vecPhoneHeight[i]
			}

			for gridAzimuthal := 0.0; gridAzimuthal <= spanAzimuthal; gridAzimuthal += angularResolution {
				for gridPolar := 0.0; gridPolar <= spanPolar; gridPolar += angularResolution {
					linalg.Equivalent(locationPhonon, locationSpeaker, 3)

					linalg.Rotation(rotationPolar, vecPhoneLength, gridPolar)
					linalg.Rotation(rotationAzimuthal, vecPhoneHeight, gridAzimuthal)

					linalg.MatrixMatrixVecMultiply(projectionPhonon, rotationPolar, rotationAzimuthal, initialPhononProjection, 3)

					atUser = false
					for !atUser {
						aParaboloid = coefficientsParaboloidX*math.Pow(projectionPhonon[0], 2) + coefficientsParaboloidY*math.Pow(projectionPhonon[1], 2)*math.Pow(math.Cos(angleParaboloid), 2) + coefficientsParaboloidY*math.Pow(projectionPhonon[2], 2)*math.Pow(math.Sin(angleParaboloid), 2) + 2.0*coefficientsParaboloidY*projectionPhonon[1]*projectionPhonon[2]*math.Cos(angleParaboloid)*math.Sin(angleParaboloid)
						bParaboloid = 2*coefficientsParaboloidX*projectionPhonon[0]*locationPhonon[0] + 2.0*coefficientsParaboloidY*projectionPhonon[1]*locationPhonon[1]*math.Pow(math.Cos(angleParaboloid), 2) + 2.0*coefficientsParaboloidY*projectionPhonon[2]*locationPhonon[2]*math.Pow(math.Sin(angleParaboloid), 2) + 2.0*coefficientsParaboloidY*projectionPhonon[1]*locationPhonon[2]*math.Cos(angleParaboloid)*math.Sin(angleParaboloid) + 2.0*coefficientsParaboloidY*projectionPhonon[2]*locationPhonon[1]*math.Cos(angleParaboloid)*math.Sin(angleParaboloid) + coefficientsParaboloidZ*projectionPhonon[1]*math.Sin(angleParaboloid) - coefficientsParaboloidZ*projectionPhonon[2]*math.Cos(angleParaboloid)
						cParaboloid = coefficientsParaboloidX*math.Pow(locationPhonon[0], 2) + coefficientsParaboloidY*math.Pow(locationPhonon[1], 2)*math.Pow(math.Cos(angleParaboloid), 2) + coefficientsParaboloidY*math.Pow(locationPhonon[2], 2)*math.Pow(math.Sin(angleParaboloid), 2) + 2.0*coefficientsParaboloidY*locationPhonon[1]*locationPhonon[2]*math.Cos(angleParaboloid)*math.Sin(angleParaboloid) + coefficientsParaboloidZ*locationPhonon[1]*math.Sin(angleParaboloid) - coefficientsParaboloidZ*locationPhonon[2]*math.Cos(angleParaboloid)
						tParaboloid = (-bParaboloid + math.Sqrt(math.Pow(bParaboloid, 2)-4*aParaboloid*cParaboloid)) / (2 * aParaboloid)
						linalg.Intersection(intersectParaboloid, locationPhonon, projectionPhonon, tParaboloid)

						tSlicingPlane = (heightSlicingPlane*math.Cos(angleSlicingPlane) + locationPhonon[1]*math.Cos(angleParaboloid)*math.Sin(angleSlicingPlane) + locationPhonon[1]*math.Sin(angleParaboloid)*math.Cos(angleSlicingPlane) - locationPhonon[2]*math.Cos(angleParaboloid)*math.Cos(angleSlicingPlane) + locationPhonon[2]*math.Sin(angleParaboloid)*math.Sin(angleSlicingPlane)) / (-projectionPhonon[1]*math.Cos(angleParaboloid)*math.Sin(angleSlicingPlane) - projectionPhonon[1]*math.Sin(angleParaboloid)*math.Cos(angleSlicingPlane) + projectionPhonon[2]*math.Cos(angleParaboloid)*math.Cos(angleSlicingPlane) - projectionPhonon[2]*math.Sin(angleParaboloid)*math.Sin(angleSlicingPlane))

						tPhone = (linalg.DotProduct(vecPhoneHeight, cornerPhone, 3) - linalg.DotProduct(vecPhoneHeight, locationPhonon, 3)) / (linalg.DotProduct(vecPhoneHeight, projectionPhonon, 3))
						linalg.Intersection(intersectPhone, locationPhonon, projectionPhonon, tPhone)

						aUser = linalg.DotProduct(projectionPhonon, projectionPhonon, 3)
						bUser = 2 * linalg.DotProduct(locationPhonon, projectionPhonon, 3)
						cUser = linalg.DotProduct(locationPhonon, locationPhonon, 3) - math.Pow(radiusUser, 2)
						tUser = (-bUser + math.Sqrt(math.Pow(bUser, 2)-4*aUser*cUser)) / (2 * aUser)
						linalg.Intersection(intersectUser, locationPhonon, projectionPhonon, tUser)

						if tPhone > math.Pow(10.0, -6) && (tPhone < tParaboloid || tParaboloid <= math.Pow(10.0, -6)) && (tPhone < tSlicingPlane || tSlicingPlane <= math.Pow(10.0, -6)) && ((xPhone-widthPhone) <= intersectPhone[0] && intersectPhone[0] <= xPhone && yPhone <= intersectPhone[1] && intersectPhone[1] <= (yPhone+lengthPhone*math.Sin(anglePhone))) {
							linalg.Equivalent(locationPhonon, intersectPhone, 3)

							linalg.Reflect(projectionPhonon, vecPhoneHeight)

							phoneVerticies = append(phoneVerticies, locationPhonon[0]/1000, locationPhonon[1]/1000, locationPhonon[2]/1000)
						} else if tParaboloid > math.Pow(10.0, -6) && (tParaboloid < tSlicingPlane || tSlicingPlane <= math.Pow(10.0, -6)) {
							linalg.Equivalent(locationPhonon, intersectParaboloid, 3)

							vecNormal := []float64{0, 0, 0}

							vecNormal[0] = -2 * coefficientsParaboloidX * locationPhonon[0]
							vecNormal[1] = -2*coefficientsParaboloidY*math.Cos(angleParaboloid)*(locationPhonon[1]*math.Cos(angleParaboloid)+locationPhonon[2]*math.Sin(angleParaboloid)) - coefficientsParaboloidZ*math.Sin(angleParaboloid)
							vecNormal[2] = -2*coefficientsParaboloidY*math.Sin(angleParaboloid)*(locationPhonon[1]*math.Cos(angleParaboloid)+locationPhonon[2]*math.Sin(angleParaboloid)) + coefficientsParaboloidZ*math.Cos(angleParaboloid)
							linalg.Normalize(vecNormal, 3)

							linalg.Reflect(projectionPhonon, vecNormal)

							paraboloidVerticies = append(paraboloidVerticies, locationPhonon[0]/1000, locationPhonon[1]/1000, locationPhonon[2]/1000)
						} else if tSlicingPlane > math.Pow(10.0, -6) {
							atUser = true

							linalg.Equivalent(locationPhonon, intersectUser, 3)
							if locationPhonon[0] == math.Inf(-1) || locationPhonon[1] == math.Inf(-1) || locationPhonon[2] == math.Inf(-1) || locationPhonon[0] == math.Inf(1) || locationPhonon[1] == math.Inf(1) || locationPhonon[2] == math.Inf(1) || math.IsNaN(locationPhonon[0]) || math.IsNaN(locationPhonon[1]) || math.IsNaN(locationPhonon[2]) {
								fmt.Printf("User: (%g,%g,%g,%g,%g)\n", aUser, bUser, cUser, tUser, radiusUser)
								break
							}
							userVerticies = append(userVerticies, locationPhonon[0]/1000, locationPhonon[1]/1000, locationPhonon[2]/1000)
						} else {
							break
						}
					}

					if gridAzimuthal == 0 {
						break
					}
				}
			}
		}
	}
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
	theta := float64(math.Pi / 36)
	radius := float64(2.0)
	for i := int64(0); i < numVerticies; i++ {
		x := rand.Float64()*radius - radius/2
		y := rand.Float64()*radius - radius/2

		a := 0.170794
		b := a
		z := math.Pow(x, 2.0)/math.Pow(a, 2.0) + math.Pow(y, 2.0)/math.Pow(b, 2.0)

		if shape == "paraboloid" {
			verticies[3*i] = x
			verticies[3*i+1] = y*math.Cos(phi) - z*math.Sin(phi)
			verticies[3*i+2] = y*math.Sin(phi) + z*math.Cos(phi)
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

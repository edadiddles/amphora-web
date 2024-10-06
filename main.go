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
	anglePhone := phoneInput.Angle * math.Pi / 180

	speakerWidth := phoneConfig.Speaker.Width
	speakerHeight := phoneConfig.Speaker.Height
	speakerCenter := phoneConfig.Speaker.Center

	coefficientsParaboloidX := paraboloidInput.X
	coefficientsParaboloidY := paraboloidInput.Y
	coefficientsParaboloidZ := paraboloidInput.Z
	angleParaboloid := paraboloidInput.Angle * math.Pi / 180

	slicingPlaneHeight := slicingPlaneInput.Height * 10 // input in cm
	slicingPlaneAngle := slicingPlaneInput.Angle

	userRadius := userRadiusInput.Radius * 1000 // input in meters

	phoneCorner := make([]float64, 3)
	phoneCorner[0] = widthPhone / 2
	phoneCorner[1] = +(-8.0*math.Pow(widthPhone, 2.0)*coefficientsParaboloidX*coefficientsParaboloidY*math.Sin(angleParaboloid) - 6.0*math.Pow(lengthPhone, 2.0)*math.Pow(coefficientsParaboloidY, 2.0)*math.Sin(angleParaboloid) - 8.0*math.Pow(coefficientsParaboloidZ, 2.0)*math.Sin(angleParaboloid) - 16.0*lengthPhone*coefficientsParaboloidY*coefficientsParaboloidZ*math.Sin(anglePhone) - 8.0*lengthPhone*coefficientsParaboloidY*coefficientsParaboloidZ*math.Sin(2.0*angleParaboloid+anglePhone) - 4.0*math.Pow(widthPhone, 2.0)*coefficientsParaboloidX*coefficientsParaboloidY*math.Sin(angleParaboloid+2.0*anglePhone) - 4.0*math.Pow(lengthPhone, 2.0)*math.Pow(coefficientsParaboloidY, 2.0)*math.Sin(angleParaboloid+2.0*anglePhone) + 12.0*math.Pow(coefficientsParaboloidZ, 2.0)*math.Sin(angleParaboloid+2.0*anglePhone) + 4.0*math.Pow(widthPhone, 2.0)*coefficientsParaboloidX*coefficientsParaboloidY*math.Sin(3.0*angleParaboloid+2.0*anglePhone) + 4.0*math.Pow(lengthPhone, 2.0)*math.Pow(coefficientsParaboloidY, 2.0)*math.Sin(3.0*angleParaboloid+2.0*anglePhone) + 4.0*math.Pow(coefficientsParaboloidZ, 2.0)*math.Sin(3.0*angleParaboloid+2.0*anglePhone) + 8.0*lengthPhone*coefficientsParaboloidY*coefficientsParaboloidZ*math.Sin(2.0*angleParaboloid+3.0*anglePhone) + math.Pow(lengthPhone, 2.0)*math.Pow(coefficientsParaboloidY, 2.0)*math.Sin(3.0*angleParaboloid+4.0*anglePhone) - math.Pow(lengthPhone, 2.0)*math.Pow(coefficientsParaboloidY, 2.0)*math.Sin(5.0*angleParaboloid+4.0*anglePhone)) / (64.0 * coefficientsParaboloidY * coefficientsParaboloidZ * math.Pow(math.Sin(angleParaboloid+anglePhone), 2.0))
	phoneCorner[2] = -(-8.0*math.Pow(widthPhone, 2.0)*math.Cos(angleParaboloid)*coefficientsParaboloidX*coefficientsParaboloidY + 4.0*math.Pow(widthPhone, 2.0)*math.Cos(angleParaboloid+2.0*anglePhone)*coefficientsParaboloidX*coefficientsParaboloidY + 4.0*math.Pow(widthPhone, 2.0)*math.Cos(3.0*angleParaboloid+2.0*anglePhone)*coefficientsParaboloidX*coefficientsParaboloidY - 6.0*math.Pow(lengthPhone, 2.0)*math.Cos(angleParaboloid)*math.Pow(coefficientsParaboloidY, 2.0) + 4.0*math.Pow(lengthPhone, 2.0)*math.Cos(angleParaboloid+2.0*anglePhone)*math.Pow(coefficientsParaboloidY, 2.0) + 4.0*math.Pow(lengthPhone, 2.0)*math.Cos(3.0*angleParaboloid+2.0*anglePhone)*math.Pow(coefficientsParaboloidY, 2.0) - math.Pow(lengthPhone, 2.0)*math.Cos(3.0*angleParaboloid+4.0*anglePhone)*math.Pow(coefficientsParaboloidY, 2.0) - math.Pow(lengthPhone, 2.0)*math.Cos(5.0*angleParaboloid+4.0*anglePhone)*math.Pow(coefficientsParaboloidY, 2.0) + 16.0*lengthPhone*math.Cos(anglePhone)*coefficientsParaboloidY*coefficientsParaboloidZ - 8.0*lengthPhone*math.Cos(2.0*angleParaboloid+anglePhone)*coefficientsParaboloidY*coefficientsParaboloidZ - 8.0*lengthPhone*math.Cos(2.0*angleParaboloid+3.0*anglePhone)*coefficientsParaboloidY*coefficientsParaboloidZ - 8.0*math.Cos(angleParaboloid)*math.Pow(coefficientsParaboloidZ, 2.0) - 12.0*math.Cos(angleParaboloid+2.0*anglePhone)*math.Pow(coefficientsParaboloidZ, 2.0) + 4.0*math.Cos(3.0*angleParaboloid+2.0*anglePhone)*math.Pow(coefficientsParaboloidZ, 2.0)) / (64.0 * coefficientsParaboloidY * coefficientsParaboloidZ * math.Pow(math.Sin(angleParaboloid+anglePhone), 2.0))

	vecPhoneWidth := []float64{widthPhone, 0, 0}
	vecPhoneWidth = linalg.Normalize(vecPhoneWidth)

	vecPhoneLength := []float64{0, lengthPhone * math.Sin(anglePhone), lengthPhone * math.Cos(anglePhone)}
	vecPhoneLength = linalg.Normalize(vecPhoneLength)

	vecPhoneHeight := linalg.CrossProduct(vecPhoneLength, vecPhoneWidth) // should this be length X width or width X length?
	vecPhoneHeight = linalg.Normalize(vecPhoneHeight)

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

	//initialize phonon
	initialPhononLoc := make([]float64, 3)
	initialPhononProj := make([]float64, 3)
	for i := 0; i < len(initialPhononLoc); i++ {
		initialPhononLoc[i] = phoneCorner[i] - speakerCenter*vecPhoneWidth[i] + 0.5*heightPhone*vecPhoneHeight[i]
		initialPhononProj[i] = -vecPhoneLength[i]
	}

	linearResolution := resolutionInput.Linear
	angularResolution := resolutionInput.Angular
	var sum int64
	var phoneVerticies []float64
	var paraboloidVerticies []float64
	var userVerticies []float64
	for gridSpeakerWidth := -0.5 * speakerWidth; gridSpeakerWidth <= 0.5*speakerWidth; gridSpeakerWidth += linearResolution {
		for gridSpeakerHeight := -0.5 * speakerHeight; gridSpeakerHeight <= 0.5*speakerHeight; gridSpeakerHeight += linearResolution {
			speakerLoc := make([]float64, 3)
			for i := 0; i < len(speakerLoc); i++ {
				speakerLoc[i] = initialPhononLoc[i] + gridSpeakerWidth*vecPhoneWidth[i] + gridSpeakerHeight*vecPhoneHeight[i]
			}

			for gridAzimuthal := 0.0; gridAzimuthal <= spanAzimuthal; gridAzimuthal += angularResolution {
				for gridPolar := 0.0; gridPolar <= spanPolar; gridPolar += angularResolution {
					sum += 1
					phononLoc := speakerLoc

					rotationPolar := linalg.Rotation(vecPhoneLength, gridPolar)
					rotationAzimulthal := linalg.Rotation(vecPhoneHeight, gridAzimuthal)

					rotateMatrix := linalg.MatrixMultiply(rotationPolar, rotationAzimulthal)
					phononProj := linalg.MatrixVecMultiply(rotateMatrix, initialPhononProj)
					phononProj = linalg.Normalize(phononProj)
					fmt.Printf("Projection: (%g,%g,%g)\n", phononProj[0], phononProj[1], phononProj[2])

					atUser := false
					reflectCnt := 0
					for !atUser {
						reflectCnt++
						aParaboloid = coefficientsParaboloidX*math.Pow(phononProj[0], 2.0) + coefficientsParaboloidY*math.Pow(phononProj[1], 2.0)*math.Pow(math.Cos(angleParaboloid), 2.0) + coefficientsParaboloidY*math.Pow(phononProj[2], 2.0)*math.Pow(math.Sin(angleParaboloid), 2.0) + 2.0*coefficientsParaboloidY*phononProj[1]*phononProj[2]*math.Cos(angleParaboloid)*math.Sin(angleParaboloid)
						bParaboloid = 2*coefficientsParaboloidX*phononProj[0]*phononLoc[0] + 2.0*coefficientsParaboloidY*phononProj[1]*phononLoc[1]*math.Pow(math.Cos(angleParaboloid), 2.0) + 2.0*coefficientsParaboloidY*phononProj[2]*phononLoc[2]*math.Pow(math.Sin(angleParaboloid), 2.0) + 2.0*coefficientsParaboloidY*phononProj[1]*phononLoc[2]*math.Cos(angleParaboloid)*math.Sin(angleParaboloid) + 2.0*coefficientsParaboloidY*phononProj[2]*phononLoc[1]*math.Cos(angleParaboloid)*math.Sin(angleParaboloid) + coefficientsParaboloidZ*phononProj[1]*math.Sin(angleParaboloid) - coefficientsParaboloidZ*phononProj[2]*math.Cos(angleParaboloid)
						cParaboloid = coefficientsParaboloidX*math.Pow(phononLoc[0], 2.0) + coefficientsParaboloidY*math.Pow(phononLoc[1], 2.0)*math.Pow(math.Cos(angleParaboloid), 2.0) + coefficientsParaboloidY*math.Pow(phononLoc[2], 2.0)*math.Pow(math.Sin(angleParaboloid), 2.0) + 2.0*coefficientsParaboloidY*phononLoc[1]*phononLoc[2]*math.Cos(angleParaboloid)*math.Sin(angleParaboloid) + coefficientsParaboloidZ*phononLoc[1]*math.Sin(angleParaboloid) - coefficientsParaboloidZ*phononLoc[2]*math.Cos(angleParaboloid)
						tParaboloid = (-bParaboloid + math.Sqrt(math.Pow(bParaboloid, 2.0)-4*aParaboloid*cParaboloid)) / (2 * aParaboloid)
						intersectParaboloid := linalg.Intersection(phononLoc, phononProj, tParaboloid)

						tSlicingPlane = (slicingPlaneHeight*math.Cos(slicingPlaneAngle) + phononLoc[1]*math.Cos(angleParaboloid)*math.Sin(slicingPlaneAngle) + phononLoc[1]*math.Sin(angleParaboloid)*math.Cos(slicingPlaneAngle) - phononLoc[2]*math.Cos(angleParaboloid)*math.Cos(slicingPlaneAngle) + phononLoc[2]*math.Sin(angleParaboloid)*math.Sin(slicingPlaneAngle)) / (-phononProj[1]*math.Cos(angleParaboloid)*math.Sin(slicingPlaneAngle) - phononProj[1]*math.Sin(angleParaboloid)*math.Cos(slicingPlaneAngle) + phononProj[2]*math.Cos(angleParaboloid)*math.Cos(slicingPlaneAngle) - phononProj[2]*math.Sin(angleParaboloid)*math.Sin(slicingPlaneAngle))

						tPhone = (linalg.DotProduct(vecPhoneHeight, phoneCorner) - linalg.DotProduct(vecPhoneHeight, phononLoc)) / (linalg.DotProduct(vecPhoneHeight, phononProj))
						intersectPhone := linalg.Intersection(phononLoc, phononProj, tPhone)

						aUser = linalg.DotProduct(phononProj, vecPhoneLength)
						bUser = 2 * linalg.DotProduct(phononLoc, phononProj)
						cUser = linalg.DotProduct(phononLoc, phononLoc) - math.Pow(userRadius, 2.0)
						tUser = (-bUser + math.Sqrt(math.Pow(bUser, 2.0)-4*aUser*cUser)) / (2 * aUser)
						intersectUser := linalg.Intersection(phononLoc, phononProj, tUser)

						if tPhone > math.Pow(10.0, -6.0) && tPhone < tParaboloid || tParaboloid <= math.Pow(10.0, -6.0) && tPhone < tSlicingPlane || tSlicingPlane <= math.Pow(10, -6) && (phoneCorner[0]-widthPhone) <= intersectPhone[0] && intersectPhone[0] <= phoneCorner[0] && phoneCorner[1] <= intersectPhone[1] && intersectPhone[1] <= (phoneCorner[1]+lengthPhone*math.Sin(anglePhone)) {
							phononLoc = intersectPhone
							phononProj = linalg.Reflect(phononProj, vecPhoneHeight)

							if phononLoc[0] == math.Inf(-1) || phononLoc[1] == math.Inf(-1) || phononLoc[2] == math.Inf(-1) || phononLoc[0] == math.Inf(1) || phononLoc[1] == math.Inf(1) || phononLoc[2] == math.Inf(1) || math.IsNaN(phononLoc[0]) || math.IsNaN(phononLoc[1]) || math.IsNaN(phononLoc[2]) {
								fmt.Printf("Phone: (%g)\n", tPhone)
								break
							}
							phoneVerticies = append(phoneVerticies, phononLoc[0]/1000, phononLoc[1]/1000, phononLoc[2]/1000)
						} else if tParaboloid > math.Pow(10, -6) && (tParaboloid < tSlicingPlane || tSlicingPlane <= math.Pow(10, -6)) {
							phononLoc = intersectParaboloid
							vecNormal := []float64{0, 0, 0}

							vecNormal[0] = -2 * coefficientsParaboloidX * phononLoc[0]
							vecNormal[1] = -2*coefficientsParaboloidY*math.Cos(angleParaboloid)*(phononLoc[1]*math.Cos(angleParaboloid)+phononLoc[2]*math.Sin(angleParaboloid)) - coefficientsParaboloidZ*math.Sin(angleParaboloid)
							vecNormal[2] = -2*coefficientsParaboloidZ*math.Sin(angleParaboloid)*(phononLoc[1]*math.Cos(angleParaboloid)+phononLoc[2]*math.Sin(angleParaboloid)) - coefficientsParaboloidZ*math.Cos(angleParaboloid)
							vecNormal = linalg.Normalize(vecNormal)
							phononProj = linalg.Reflect(phononProj, vecNormal)

							if phononLoc[0] == math.Inf(-1) || phononLoc[1] == math.Inf(-1) || phononLoc[2] == math.Inf(-1) || phononLoc[0] == math.Inf(1) || phononLoc[1] == math.Inf(1) || phononLoc[2] == math.Inf(1) || math.IsNaN(phononLoc[0]) || math.IsNaN(phononLoc[1]) || math.IsNaN(phononLoc[2]) {
								fmt.Printf("Paraboloid: (%g,%g,%g, %g)\n", aParaboloid, bParaboloid, cParaboloid, tParaboloid)
								break
							}
							paraboloidVerticies = append(paraboloidVerticies, phononLoc[0]/1000, phononLoc[1]/1000, phononLoc[2]/1000)
						} else if tSlicingPlane > math.Pow(10, -6) {
							atUser = true
							phononLoc = intersectUser

							if phononLoc[0] == math.Inf(-1) || phononLoc[1] == math.Inf(-1) || phononLoc[2] == math.Inf(-1) || phononLoc[0] == math.Inf(1) || phononLoc[1] == math.Inf(1) || phononLoc[2] == math.Inf(1) || math.IsNaN(phononLoc[0]) || math.IsNaN(phononLoc[1]) || math.IsNaN(phononLoc[2]) {
								fmt.Printf("User: (%g,%g,%g,%g)\n", aUser, bUser, cUser, tUser)
								break
							}
							userVerticies = append(userVerticies, phononLoc[0]/1000, phononLoc[1]/1000, phononLoc[2]/1000)
						} else {
							fmt.Printf("tPhone: %g; tParaboloid: %g; tSlicingPlane: %g; tUser: %g;\nIntersection phone: (%g,%g,%g)\nPhone corner: (%g,%g,%g)\n", tPhone, tParaboloid, tSlicingPlane, tUser, intersectPhone[0], intersectPhone[1], intersectPhone[2], phoneCorner[0], phoneCorner[1], phoneCorner[2])
							break
						}

						fmt.Printf("Projection: (%g,%g,%g)\n", phononProj[0], phononProj[1], phononProj[2])
						if reflectCnt > 100 {
							fmt.Println("max reflections exceeded")
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
	// paraboloidVerticies = generateVerticies(10000, "paraboloid")
	// phoneVerticies = nil //generateVerticies(10000, "phone")
	// userVerticies = nil  //generateVerticies(10000, "userSphere")
	// paraboloidVerticies = append(paraboloidVerticies, generateVerticies(10000, "paraboloid")...)
	// phoneVerticies = append(phoneVerticies, generateVerticies(10000, "phone")...)
	// userVerticies = append(userVerticies, generateVerticies(10000, "userSphere")...)
	fmt.Printf("#Phone: %d\n#Paraboloid: %d\n#User: %d\n", len(phoneVerticies), len(paraboloidVerticies), len(userVerticies))
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

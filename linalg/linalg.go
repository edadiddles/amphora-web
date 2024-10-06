package linalg

import (
	"math"
)

func DotProduct(a []float64, b []float64) float64 {
	sum := 0.0
	for i := 0; i < len(a); i++ {
		sum += a[i] * b[i]
	}

	return sum
}

func CrossProduct(a []float64, b []float64) []float64 {
	product := []float64{a[1]*b[2] - a[2]*b[1], a[2]*b[0] - a[0]*b[2], a[0]*b[1] - a[1]*b[0]}
	return product
}

func Rotation(axisOfRotation []float64, rotationAngle float64) [][]float64 {
	rotationMatrix := [][]float64{
		{0, 0, 0},
		{0, 0, 0},
		{0, 0, 0},
	}

	rotationMatrix[0][0] = math.Cos(rotationAngle) + axisOfRotation[0]*axisOfRotation[0]*(1-math.Cos(rotationAngle))
	rotationMatrix[1][1] = math.Cos(rotationAngle) + axisOfRotation[1]*axisOfRotation[1]*(1-math.Cos(rotationAngle))
	rotationMatrix[2][2] = math.Cos(rotationAngle) + axisOfRotation[2]*axisOfRotation[2]*(1-math.Cos(rotationAngle))

	rotationMatrix[0][1] = -axisOfRotation[2]*math.Sin(rotationAngle) + axisOfRotation[0]*axisOfRotation[1]*(1-math.Cos(rotationAngle))
	rotationMatrix[0][2] = axisOfRotation[1]*math.Sin(rotationAngle) + axisOfRotation[0]*axisOfRotation[2]*(1-math.Cos(rotationAngle))

	rotationMatrix[1][0] = axisOfRotation[2]*math.Sin(rotationAngle) + axisOfRotation[0]*axisOfRotation[1]*(1-math.Cos(rotationAngle))
	rotationMatrix[1][2] = -axisOfRotation[0]*math.Sin(rotationAngle) + axisOfRotation[1]*axisOfRotation[2]*(1-math.Cos(rotationAngle))

	rotationMatrix[2][0] = -axisOfRotation[1]*math.Sin(rotationAngle) + axisOfRotation[0]*axisOfRotation[2]*(1-math.Cos(rotationAngle))
	rotationMatrix[2][1] = axisOfRotation[0]*math.Sin(rotationAngle) + axisOfRotation[1]*axisOfRotation[2]*(1-math.Cos(rotationAngle))

	return rotationMatrix
}

func MatrixMultiply(matrix1 [][]float64, matrix2 [][]float64) [][]float64 {
	outputMatrix := [][]float64{
		{0, 0, 0},
		{0, 0, 0},
		{0, 0, 0},
	}

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			outputMatrix[i][j] = 0
			for k := 0; k < 3; k++ {
				outputMatrix[i][j] += matrix1[i][k] * matrix2[k][j]
			}
		}
	}

	return outputMatrix
}

func MatrixVecMultiply(matrix [][]float64, inpVec []float64) []float64 {
	outVec := make([]float64, 3)

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if j == 0 {
				outVec[i] = 0
			}

			outVec[i] += matrix[i][j] * inpVec[j]

		}
	}

	return outVec
}

func Intersection(phononLoc []float64, phononProj []float64, lengthToIntersect float64) []float64 {
	intersection := make([]float64, 3)
	for i := 0; i < 3; i++ {
		intersection[i] = phononLoc[i] + lengthToIntersect*phononProj[i]
	}

	return intersection
}

func Reflect(vec1 []float64, vec2 []float64) []float64 {
	vec1 = Normalize(vec1)
	vec2 = Normalize(vec2)

	angleReflect := math.Acos(-DotProduct(vec2, vec1))
	reflectVec := CrossProduct(vec2, vec1)
	reflectVec = Normalize(reflectVec)
	reflectionRotation := Rotation(reflectVec, angleReflect)

	outVec := MatrixVecMultiply(reflectionRotation, vec2)
	return outVec
}

func Normalize(vec []float64) []float64 {
	outVec := make([]float64, 3)
	vecLength := math.Sqrt(DotProduct(vec, vec))
	for i := 0; i < 3; i++ {
		outVec[i] = vec[i] / vecLength
	}
	return outVec
}

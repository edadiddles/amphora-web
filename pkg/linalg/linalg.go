package linalg

import (
	"math"
)

func Equivalent(outVec []float64, inVec []float64, n int) {
	for i := 0; i < n; i++ {
		outVec[i] = inVec[i]
	}
}

func DotProduct(a []float64, b []float64, n int) float64 {
	sum := 0.0
	for i := 0; i < n; i++ {
		sum += a[i] * b[i]
	}

	return sum
}

func CrossProduct(product []float64, a []float64, b []float64, n int) {
	for i := 0; i < n; i++ {
		product[i] = a[(i+1)%n]*b[(i+2)%n] - a[(i+2)%n]*b[(i+1)%n]
	}
}

func Rotation(rotationMatrix [][]float64, axisOfRotation []float64, rotationAngle float64) {

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

func MatrixMultiply(outputMatrix [][]float64, matrix1 [][]float64, matrix2 [][]float64, n int) {
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			outputMatrix[i][j] = 0
			for k := 0; k < n; k++ {
				outputMatrix[i][j] = outputMatrix[i][j] + matrix1[i][k]*matrix2[k][j]
			}
		}
	}
}

func MatrixVecMultiply(outVec []float64, matrix [][]float64, inpVec []float64, n int) {
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if j == 0 {
				outVec[i] = 0
			}

			outVec[i] = outVec[i] + matrix[i][j]*inpVec[j]

		}
	}
}

func MatrixMatrixVecMultiply(outVec []float64, matrix1 [][]float64, matrix2 [][]float64, vec []float64, n int) {
	intermediateMatrix := [][]float64{
		{0, 0, 0},
		{0, 0, 0},
		{0, 0, 0},
	}

	MatrixMultiply(intermediateMatrix, matrix1, matrix2, n)
	MatrixVecMultiply(outVec, intermediateMatrix, vec, n)
}

func Intersection(intersection []float64, phononLoc []float64, phononProj []float64, lengthToIntersect float64) {
	for i := 0; i < 3; i++ {
		intersection[i] = phononLoc[i] + lengthToIntersect*phononProj[i]
	}
}

func Reflect(vec1 []float64, vec2 []float64) {
	reflectVec := []float64{0, 0, 0}
	reflectRot := [][]float64{
		{0, 0, 0},
		{0, 0, 0},
		{0, 0, 0},
	}

	angleReflect := math.Acos(-DotProduct(vec2, vec1, 3))

	CrossProduct(reflectVec, vec2, vec1, 3)
	Normalize(reflectVec, 3)

	Rotation(reflectRot, reflectVec, angleReflect)

	MatrixVecMultiply(vec1, reflectRot, vec2, 3)
}

func Normalize(vec []float64, n int) {
	vecLength := math.Sqrt(DotProduct(vec, vec, 3))
	for i := 0; i < n; i++ {
		vec[i] = vec[i] / vecLength
	}
}

package linalg

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

func TestDotProduct(t *testing.T) {
	v1 := []float64{1, 0, 0}
	v2 := []float64{1, 1, 0}

	p := DotProduct(v1, v2)

	expect := float64(1)
	if p != expect {
		t.Fatalf("expected %g, received %g", expect, p)
	}
}

func TestCrossProduct(t *testing.T) {
	v1 := []float64{1, 0, 4}
	v2 := []float64{3, 4, 2}

	p := CrossProduct(v1, v2)

	expect := []float64{-16, 10, 4}
	if !reflect.DeepEqual(expect, p) {
		t.Fatalf("expected (%g, %g, %g), received (%g, %g, %g)", expect[0], expect[1], expect[2], p[0], p[1], p[2])
	}
}

func TestNormalize(t *testing.T) {
	vec := []float64{0, 4, 0}

	v := Normalize(vec)

	expect := []float64{0, 1, 0}
	if !reflect.DeepEqual(expect, v) {
		t.Fatalf("expected (%g, %g, %g), received (%g, %g, %g)", expect[0], expect[1], expect[2], v[0], v[1], v[2])
	}
}

func TestRotation(t *testing.T) {
	axisOfRotation := []float64{1, 1, 1}
	rotationAngle := float64(math.Pi / 3)

	m := Rotation(axisOfRotation, rotationAngle)

	expect := [][]float64{
		{0, 0, 0},
		{0, 0, 0},
		{0, 0, 0},
	}

	if !reflect.DeepEqual(expect, m) {
		t.Fatalf("not the same")
	}
}

func TestMatrixMultiply(t *testing.T) {
	m1 := [][]float64{
		{5, 10, 2},
		{4, 7, 6},
		{9, 9, 4},
	}
	m2 := [][]float64{
		{7, 5, 11},
		{2, 5, 3},
		{1, 9, 2},
	}

	expect := [][]float64{
		{57, 93, 89},
		{48, 109, 77},
		{85, 126, 134},
	}
	m := MatrixMultiply(m1, m2)

	if !reflect.DeepEqual(expect, m) {
		t.Fatalf("not the same")
	}
}

func TestMatrixVecMultiply(t *testing.T) {
	mat := [][]float64{
		{5, 8, 1},
		{3, 7, 4},
		{7, 1, 3},
	}
	vec := []float64{4, 2, 2}

	expect := []float64{38, 34, 36}
	m := MatrixVecMultiply(mat, vec)

	if !reflect.DeepEqual(expect, m) {
		t.Fatalf("not the same")
	}
}

func TestIntersection(t *testing.T) {
	v1 := []float64{1, 5, 3}
	v2 := []float64{3, 7, 9}
	s := float64(12)

	expect := []float64{37, 89, 111}
	v := Intersection(v1, v2, s)

	if !reflect.DeepEqual(expect, v) {
		t.Fatalf("not the same")
	}
}

func TestReflect(t *testing.T) {
	v1 := []float64{4, 3, 3}
	v2 := []float64{2, 7, 3}

	expect := []float64{0, 0, 0}
	v := Reflect(v1, v2)

	vec1 := Normalize(v1)
	vec2 := Normalize(v2)

	angleReflect := math.Acos(-DotProduct(vec2, vec1))
	fmt.Println(angleReflect)

	if !reflect.DeepEqual(expect, v) {
		t.Fatalf("not the same")
	}
}

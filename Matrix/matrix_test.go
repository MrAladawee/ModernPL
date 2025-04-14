package main

import (
	"math"
	"testing"
)

const epsilon = 1e-6

func floatsAlmostEqual(a, b float64) bool {
	return math.Abs(a-b) < epsilon
}

func TestGenerateMatrixSize(t *testing.T) {
	n := 10
	matrix := generateMatrix(n)
	if len(matrix) != n {
		t.Errorf("Ожидаемое число строк: %d, получено %d", n, len(matrix))
	}
	for _, row := range matrix {
		if len(row) != n {
			t.Errorf("Ожидаемое число столбцов: %d, получено %d", n, len(row))
		}
	}
}

func TestDeterminantIdentityMatrix(t *testing.T) {
	n := 4
	identity := make([][]float64, n)
	for i := range identity {
		identity[i] = make([]float64, n)
		identity[i][i] = 1
	}
	det := determinant(identity, n)
	if !floatsAlmostEqual(det, 1.0) {
		t.Errorf("Ожидаемое значение от единичной матрицы - 1, получено %f", det)
	}
}

func TestDeterminantZeroMatrix(t *testing.T) {
	n := 5
	zeroMatrix := make([][]float64, n)
	for i := range zeroMatrix {
		zeroMatrix[i] = make([]float64, n)
	}
	det := determinant(zeroMatrix, n)
	if !floatsAlmostEqual(det, 0.0) {
		t.Errorf("Ожидаемое значение от нулевой матрицы - 0, получено %f", det)
	}
}

func TestDeterminantDiagonalMatrix(t *testing.T) {
	n := 3
	diag := [][]float64{
		{2, 0, 0},
		{0, 3, 0},
		{0, 0, -4},
	}
	expected := 2 * 3 * -4
	det := determinant(diag, n)
	if !floatsAlmostEqual(det, float64(expected)) {
		t.Errorf("Ожидаемое значение %d, полученное %f", expected, det)
	}
}

func TestDeterminantKnownMatrix(t *testing.T) {
	matrix := [][]float64{
		{1, 2, 3},
		{0, 1, 4},
		{5, 6, 0},
	}
	// Определитель должен быть 1*1*0 + 2*4*5 + 3*0*6 - 3*1*5 - 2*0*0 - 1*4*6 = 0 + 40 + 0 - 15 - 0 - 24 = 1
	expected := 1.0
	det := determinant(matrix, 3)
	if !floatsAlmostEqual(det, expected) {
		t.Errorf("Ожидаемое значение %f, полученное %f", expected, det)
	}
}

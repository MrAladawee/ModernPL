package main

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"
)

func generateMatrix(n int) [][]float64 {
	matrix := make([][]float64, n)
	for i := range matrix {
		matrix[i] = make([]float64, n)
		for j := range matrix[i] {
			// Случайные числа в диапазоне от -10 до 10
			matrix[i][j] = rand.Float64()*20 - 10
		}
	}
	return matrix
}

func printMatrix(matrix [][]float64, n int) {
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			fmt.Printf("%8.4f ", matrix[i][j])
		}
		fmt.Println()
	}
}

func determinant(matrix [][]float64, n int) float64 {

	det := 1.0
	sign := 1.0

	// Поиск опорного элемента столбца i
	for i := 0; i < n; i++ {
		pivot := matrix[i][i]
		pivotRow := i
		for j := i; j < n; j++ {
			if matrix[j][i] != 0 {
				pivot = matrix[j][i]
				pivotRow = j
				break
			}
		}

		// Если опорный элемент равен нулю, определитель равен 0.
		if pivot == 0 {
			return 0
		}

		if pivotRow != i {
			matrix[i], matrix[pivotRow] = matrix[pivotRow], matrix[i]
			sign = -sign // обмен строк меняет знак определителя
		}

		var wg sync.WaitGroup
		for j := i + 1; j < n; j++ {
			wg.Add(1)
			go func(j int) {
				defer wg.Done()
				factor := matrix[j][i] / matrix[i][i]
				for k := i; k < n; k++ {
					matrix[j][k] -= factor * matrix[i][k]
				}
			}(j)
		}
		wg.Wait()
	}

	// Подсчёт определителя после Гаусса
	for i := 0; i < n; i++ {
		det *= matrix[i][i]
	}
	return det * sign
}

func main() {
	rand.Seed(time.Now().UnixNano())

	var n int
	fmt.Print("Введите размер матрицы (от 5 до 500): ")
	_, err := fmt.Scan(&n)
	if err != nil {
		fmt.Println("Ошибка ввода:", err)
		os.Exit(1)
	}
	if n < 5 || n > 500 {
		fmt.Println("Размер матрицы должен быть в диапазоне от 5 до 500.")
		os.Exit(1)
	}

	matrix := generateMatrix(n)
	fmt.Printf("Сгенерирована матрица размером %dx%d\n", n, n)

	fmt.Println("Исходная матрица:")
	printMatrix(matrix, n)

	start := time.Now()
	det := determinant(matrix, n)
	elapsed := time.Since(start)
	fmt.Printf("\nОпределитель матрицы: %f\n", det)
	fmt.Printf("Время вычисления: %s\n", elapsed)
}

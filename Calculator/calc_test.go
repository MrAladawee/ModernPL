package main

import (
	"fmt"
	"math"
	"testing"
)

const epsilon = 1e-9

func TestCalculate(t *testing.T) {
	tests := []struct {
		expr    string
		want    float64
		wantErr error
	}{
		// Базовые операции
		{"2 + 3", 5, nil},
		{"10 - 4", 6, nil},
		{"6 * 7", 42, nil},
		{"20 / 5", 4, nil},
		{"3 + 4 * 2", 11, nil},
		{"(3 + 4) * 2", 14, nil},
		{"10 / 3", 3.3333333333333335, nil},

		// Унарные минусы
		{"-5 + 3", -2, nil},
		{"3 * (-2)", -6, nil},
		{"-(-4)", 4, nil},
		{"-3 + (-4)", -7, nil},
		{"-0", 0, nil},

		// Скобки
		{"(2 + 3) * (4 - 1)", 15, nil},
		{"((15 / (7 - 2)) - 3) * 2", 0, nil},
		{"1 + (2 * (3 + (4 / 2)))", 11, nil},

		// Ошибки
		{"10 / 0", 0, fmt.Errorf("division by zero")},
		{"2 + a", 0, fmt.Errorf("unknown character: a")},
		{"3.14.15", 0, fmt.Errorf("invalid number at position 4")},
		{"(2 + 3", 0, fmt.Errorf("mismatched parentheses")},

		// Числа с плавающей точкой
		{"2.5 + 3.5", 6, nil},
		{"10.2 / 2", 5.1, nil},
		{"0.1 + 0.2", 0.3, nil},
	}

	for _, tc := range tests {
		got, err := Calculate(tc.expr)

		// Проверка ошибок
		if (err != nil) != (tc.wantErr != nil) {
			t.Errorf("%s: error = %v, wantErr %v", tc.expr, err, tc.wantErr)
			continue
		}

		// Сравнение текста ошибок
		if err != nil && tc.wantErr != nil && err.Error() != tc.wantErr.Error() {
			t.Errorf("%s: error message = %v, want %v", tc.expr, err.Error(), tc.wantErr.Error())
			continue
		}

		// Проверка числовых результатов
		if err == nil && math.Abs(got-tc.want) > epsilon {
			t.Errorf("%s: got %v, want %v", tc.expr, got, tc.want)
		}
	}
}

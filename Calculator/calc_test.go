package main

import (
	"reflect"
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"3+4", []string{"3", "+", "4"}},
		{"(1 + 2) * 3", []string{"(", "1", "+", "2", ")", "*", "3"}},
		{"10 / (5 - 3)", []string{"10", "/", "(", "5", "-", "3", ")"}},
		{" 3 + 4.5 ", []string{"3", "+", "4.5"}},
	}

	for _, tc := range tests {
		result := tokenize(tc.input)
		if !reflect.DeepEqual(result, tc.expected) {
			t.Errorf("tokenize(%q) = %v, ожидалось %v", tc.input, result, tc.expected)
		}
	}
}

func TestInfixToPostfix(t *testing.T) {
	tests := []struct {
		input    []string
		expected []string
		hasError bool
	}{
		{[]string{"3", "+", "4"}, []string{"3", "4", "+"}, false},
		{[]string{"(", "1", "+", "2", ")", "*", "3"}, []string{"1", "2", "+", "3", "*"}, false},
		{[]string{"10", "/", "(", "5", "-", "3", ")"}, []string{"10", "5", "3", "-", "/"}, false},
		{[]string{"(", "3", "+", "4"}, nil, true}, // Несовпадение скобок
		{[]string{"3", "+", "@"}, nil, true},      // Неизвестный символ
	}

	for _, tt := range tests {
		result, err := infixToPostfix(tt.input)
		if (err != nil) != tt.hasError {
			t.Errorf("infixToPostfix(%v) error = %v, expected error: %v", tt.input, err, tt.hasError)
		}
		if err == nil && !reflect.DeepEqual(result, tt.expected) {
			t.Errorf("infixToPostfix(%v) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}

func TestEvaluatePostfix(t *testing.T) {
	tests := []struct {
		input    []string
		expected float64
		hasError bool
	}{
		{[]string{"3", "4", "+"}, 7.0, false},
		{[]string{"10", "5", "/", "2", "*"}, 4.0, false},
		{[]string{"2", "3", "+", "4", "*"}, 20.0, false},
		{[]string{"3", "0", "/"}, 0, true}, // Деление на ноль
		{[]string{"3", "+"}, 0, true},      // Недостаточно операндов
	}

	for _, tt := range tests {
		result, err := evaluatePostfix(tt.input)
		if (err != nil) != tt.hasError {
			t.Errorf("evaluatePostfix(%v) error = %v, expected error: %v", tt.input, err, tt.hasError)
		}
		if err == nil && result != tt.expected {
			t.Errorf("evaluatePostfix(%v) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}

func TestFullEvaluation(t *testing.T) {
	tests := []struct {
		expr     string
		expected float64
		hasError bool
	}{
		{"3 + 4", 7.0, false},
		{"(1 + 2) * 3", 9.0, false},
		{"10 / (5 - 3)", 5.0, false},
		{"3 + (4 * 2 - ( 3 * 4 ) / 2) - 7 / 2", 1.5, false},
		{"3 +", 0, true},         // Ошибка
		{"(3 + 2", 0, true},      // Скобки
		{"3 / 0", 0, true},       // Деление на ноль
		{"3 + unknown", 0, true}, // Неизвестный токен
	}

	for _, tt := range tests {
		tokens := tokenize(tt.expr)
		postfix, err := infixToPostfix(tokens)
		if err != nil && !tt.hasError {
			t.Errorf("infixToPostfix(%q) unexpected error: %v", tt.expr, err)
			continue
		}
		if err == nil {
			result, err := evaluatePostfix(postfix)
			if (err != nil) != tt.hasError {
				t.Errorf("evaluatePostfix(%q) error = %v, expected error: %v", tt.expr, err, tt.hasError)
			}
			if err == nil && result != tt.expected {
				t.Errorf("evaluate(%q) = %v, expected %v", tt.expr, result, tt.expected)
			}
		}
	}
}

package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

var priority = map[string]int{
	"+": 1,
	"-": 1,
	"*": 2,
	"/": 2,
}

func isOperator(symbol string) bool {
	return symbol == "+" || symbol == "-" || symbol == "*" || symbol == "/"
}

// tokenize разбивает строку выражения на отдельные токены.
// Например, "3+(4*2)-7/1" преобразуется в: ["3", "+", "(", "4", "*", "2", ")", "-", "7", "/", "1"].
func tokenize(expr string) []string {

	var tokens []string        // По сути стек для операторов
	var number strings.Builder // tmp для чисел

	for _, ch := range expr {
		if unicode.IsDigit(ch) || ch == '.' {
			number.WriteRune(ch) // Процесс накопления числа
		} else {
			if number.Len() > 0 {
				tokens = append(tokens, number.String())
				number.Reset()
			}
			if unicode.IsSpace(ch) {
				continue
			}
			tokens = append(tokens, string(ch))
		}
	}

	// если осталось число
	if number.Len() > 0 {
		tokens = append(tokens, number.String())
	}

	return tokens

}

// infixToPostfix преобразует список токенов из инфиксной записи в постфиксную (обратную польскую запись)
func infixToPostfix(tokens []string) ([]string, error) {

	var output []string  // ОПЗ
	var opStack []string // стэк-операторов

	for _, token := range tokens {

		if _, err := strconv.ParseFloat(token, 64); err == nil {
			output = append(output, token) // если число - добавляем

		} else if token == "(" {
			opStack = append(opStack, token)

		} else if token == ")" {

			// Извлекаем операторы до открывающей скобки справа налево
			found := false
			for len(opStack) > 0 {
				top := opStack[len(opStack)-1]
				opStack = opStack[:len(opStack)-1]
				if top == "(" {
					found = true
					break
				}
				output = append(output, top)
			}

			if !found {
				return nil, fmt.Errorf("не совпадают скобки") // Обработаем ошибку на тупого со скобками
			}

		} else if isOperator(token) {
			// Для оператора проверяем приоритет и выталкиваем операторы из стека
			for len(opStack) > 0 {
				top := opStack[len(opStack)-1]
				if isOperator(top) && priority[top] >= priority[token] {
					opStack = opStack[:len(opStack)-1]
					output = append(output, top)
				} else {
					break
				}
			}
			opStack = append(opStack, token)
		} else {
			return nil, fmt.Errorf("неизвестный токен: %s", token)
		}
	}

	// Добавляем оставшиеся операторы в выходной список
	for len(opStack) > 0 {
		top := opStack[len(opStack)-1]
		opStack = opStack[:len(opStack)-1]
		if top == "(" || top == ")" {
			return nil, fmt.Errorf("не совпадают скобки") // Снова ошибка на тупого
		}
		output = append(output, top)
	}

	return output, nil
}

// evaluatePostfix вычисляет значение выражения, заданного в постфиксной записи.
func evaluatePostfix(postfix []string) (float64, error) {

	var stack []float64

	for _, token := range postfix {
		if num, err := strconv.ParseFloat(token, 64); err == nil {

			// Если токен число, кладём его в стек.
			stack = append(stack, num)
			
		} else if isOperator(token) {

			// Проверка наличия двух операндов
			if len(stack) < 2 {
				return 0, fmt.Errorf("недостаточно операндов (чисел) для оператора %s", token)
			}

			// Извлекаем два числа (правый операнд извлекается первым, чтобы сразу подчищать стэк)
			right := stack[len(stack)-1]
			left := stack[len(stack)-2]
			stack = stack[:len(stack)-2]

			var result float64
			switch token {
			case "+":
				result = left + right
			case "-":
				result = left - right
			case "*":
				result = left * right
			case "/":
				if right == 0 {
					return 0, fmt.Errorf("деление на ноль")
				}
				result = left / right
			}
			// Результат помещаем обратно в стек.
			stack = append(stack, result)
		} else {
			return 0, fmt.Errorf("неизвестный токен: %s", token)
		}
	}

	// После вычисления в стеке должен остаться ровно один элемент — результат.
	if len(stack) != 1 {
		return 0, fmt.Errorf("ошибка вычисления выражения")
	}
	return stack[0], nil
}

func main() {
	expression := "3 + (4 * 2 - ( 3 * 4 - 2) / 2) - 7 / 2" // Правильный ответ: 1.5
	fmt.Println("Аходное выражение:", expression)

	tokens := tokenize(expression)
	fmt.Println("Токены:", tokens)

	postfix, err := infixToPostfix(tokens)
	if err != nil {
		fmt.Println("Ошибка при преобразовании выражения:", err)
		return
	}
	fmt.Println("Постфиксная запись:", postfix)

	result, err := evaluatePostfix(postfix)
	if err != nil {
		fmt.Println("Ошибка при вычислении выражения:", err)
		return
	}
	fmt.Printf("Результат: %.2f\n", result)

}

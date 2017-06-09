package wine

import (
	"strings"
)

func plus(a, b int) int {
	return a + b
}

func minus(a, b int) int {
	return a - b
}

func multiple(a, b int) int {
	return a * b
}

func divide(a, b int) int {
	return a / b
}

func join(strs []string, sep string) string {
	return strings.Join(strs, sep)
}

package template

import (
	"html/template"
	"strings"
)

var FuncMap = template.FuncMap{
	"plus":     Plus,
	"minus":    Minus,
	"multiple": Multiple,
	"divide":   Divide,
	"join":     Join,
}

func Plus(a, b int) int {
	return a + b
}

func Minus(a, b int) int {
	return a - b
}

func Multiple(a, b int) int {
	return a * b
}

func Divide(a, b int) int {
	return a / b
}

func Join(strs []string, sep string) string {
	return strings.Join(strs, sep)
}

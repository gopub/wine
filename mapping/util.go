package mapping

func CamelToSnake(s string) string {
	snake := make([]rune, 0, len(s)+1)
	flag := false
	k := 'a' - 'A'
	for i, c := range s {
		if c >= 'A' && c <= 'Z' {
			if !flag {
				flag = true
				if i > 0 {
					snake = append(snake, '_')
				}
			}
			snake = append(snake, c+k)
		} else {
			flag = false
			snake = append(snake, c)
		}
	}
	return string(snake)
}

func SnakeToCamel(s string) string {
	camel := make([]rune, 0, len(s)+1)
	flag := false
	k := 'A' - 'a'
	for _, c := range s {
		if c == '_' {
			flag = true
			continue
		}

		if flag {
			flag = false
			if c >= 'a' && c <= 'z' {
				camel = append(camel, c+k)
				continue
			}
		}
		camel = append(camel, c)
	}
	return string(camel)
}

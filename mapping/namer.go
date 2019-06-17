package mapping

type Namer interface {
	Name(srcName string) (dstName string)
}

type NameFunc func(srcName string) (dstName string)

func (f NameFunc) Name(srcName string) (dstName string) {
	return f(srcName)
}

func MapNamer(srcToDst map[string]string) Namer {
	return NameFunc(func(srcName string) (dstName string) {
		return srcToDst[srcName]
	})
}

var SnakeToCamelNamer NameFunc = func(snakeSrcName string) (camelDstName string) {
	return SnakeToCamel(snakeSrcName)
}

var CamelToSnakeNamer NameFunc = func(camelSrcName string) (snakeDstName string) {
	return CamelToSnake(camelSrcName)
}

var EqaulNamer NameFunc = func(srcName string) (dstName string) {
	return srcName
}

var defaultNamer Namer = EqaulNamer

func SetDefaultNamer(namer Namer) {
	defaultNamer = namer
}

func DefaultNamer() Namer {
	return defaultNamer
}

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

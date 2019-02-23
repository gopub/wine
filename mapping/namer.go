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

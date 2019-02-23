package mapping

type Transformer interface {
	Transform(v interface{}) (result interface{}, err error)
}

type TransformFunc func(i interface{}) (result interface{}, err error)

func (p TransformFunc) Transform(i interface{}) (result interface{}, err error) {
	return p(i)
}

var nameToTransformer = map[string]Transformer{}

func SetTransformer(name string, t Transformer) {
	nameToTransformer[name] = t
}

func GetTransformer(name string) Transformer {
	return nameToTransformer[name]
}

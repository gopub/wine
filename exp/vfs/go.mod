module github.com/gopub/wine/exp/vfs

go 1.15

require (
	github.com/gabriel-vasile/mimetype v1.1.2
	github.com/google/uuid v1.1.3
	github.com/gopub/conv v0.3.27
	github.com/gopub/errors v0.1.6
	github.com/gopub/log v1.2.3
	github.com/gopub/sql v1.4.15
	github.com/gopub/types v0.2.25
	github.com/gopub/wine v1.36.2
	github.com/gopub/wine/httpvalue v0.1.0
	github.com/stretchr/testify v1.6.1
)

replace github.com/gopub/wine/httpvalue => ../../httpvalue

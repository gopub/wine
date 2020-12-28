module github.com/gopub/wine/exp/vfs

go 1.15

require (
	github.com/gabriel-vasile/mimetype v1.1.2
	github.com/google/uuid v1.1.2
	github.com/gopub/conv v0.3.27
	github.com/gopub/errors v0.1.6
	github.com/gopub/log v1.2.3
	github.com/gopub/sql v1.4.15
	github.com/gopub/types v0.2.24
	github.com/gopub/wine/httpvalue v0.0.0-20201228214056-96ee5bb82550
)

replace (
	github.com/gopub/wine/httpvalue => ../../httpvalue
)
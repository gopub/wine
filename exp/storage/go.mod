module github.com/gopub/wine/exp/storage

go 1.15

require (
	github.com/disintegration/imaging v1.6.2
	github.com/gopub/errors v0.1.6
	github.com/gopub/wine v1.36.4
	github.com/gopub/wine/httpvalue v0.1.0
	golang.org/x/image v0.0.0-20191009234506-e7c1f5e7dbb8
)

replace (
	github.com/gopub/wine => ../../
	github.com/gopub/wine/httpvalue => ../../httpvalue
	github.com/gopub/wine/router => ../../router
)

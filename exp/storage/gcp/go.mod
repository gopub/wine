module github.com/gopub/wine/exp/storage/gcp

go 1.15

require (
	cloud.google.com/go/storage v1.12.0
	github.com/gopub/log v1.2.3
	github.com/gopub/wine/exp/storage v0.1.2
)

replace (
	github.com/gopub/wine => ../../../
	github.com/gopub/wine/exp/storage => ../
	github.com/gopub/wine/httpvalue => ../../../httpvalue
	github.com/gopub/wine/router => ../../../router
)

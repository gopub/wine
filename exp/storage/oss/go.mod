module github.com/gopub/wine/exp/storage/oss

go 1.15

require (
	github.com/aliyun/aliyun-oss-go-sdk v2.1.5+incompatible
	github.com/gopub/wine v1.36.0
	github.com/gopub/wine/exp/storage v0.1.0
)

replace (
	github.com/gopub/wine => ../../../
	github.com/gopub/wine/exp/storage => ../
	github.com/gopub/wine/httpvalue => ../../../httpvalue
	github.com/gopub/wine/router => ../../../router
)

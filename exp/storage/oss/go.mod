module github.com/gopub/wine/exp/storage/oss

go 1.15

require (
	github.com/aliyun/aliyun-oss-go-sdk v2.1.6+incompatible
	github.com/gopub/errors v0.1.7
	github.com/gopub/mapper v1.0.11 // indirect
	github.com/gopub/wine v1.38.2
	github.com/gopub/wine/exp/storage v0.1.6
	github.com/gopub/wine/urlutil v0.1.0
	golang.org/x/time v0.0.0-20201208040808-7e3f01d25324 // indirect
)

//replace (
//	github.com/gopub/wine => ../../../
//	github.com/gopub/wine/exp/storage => ../
//	github.com/gopub/wine/httpvalue => ../../../httpvalue
//	github.com/gopub/wine/router => ../../../router
//)

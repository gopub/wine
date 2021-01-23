module github.com/gopub/wine/exp/storage/s3

go 1.15

require (
	github.com/aws/aws-sdk-go v1.36.31
	github.com/gopub/wine v1.38.2 // indirect
	github.com/gopub/wine/exp/storage v0.1.6
)

//replace (
//	github.com/gopub/wine => ../../../
//	github.com/gopub/wine/exp/storage => ../
//	github.com/gopub/wine/httpvalue => ../../../httpvalue
//	github.com/gopub/wine/router => ../../../router
//)

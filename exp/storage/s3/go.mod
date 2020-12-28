module github.com/gopub/wine/exp/storage/s3

go 1.15

require (
	github.com/aws/aws-sdk-go v1.36.16
	github.com/gopub/wine/exp/storage v0.1.0
)

replace (
	github.com/gopub/wine => ../../../
	github.com/gopub/wine/exp/storage => ../
	github.com/gopub/wine/httpvalue => ../../../httpvalue
	github.com/gopub/wine/router => ../../../router
)

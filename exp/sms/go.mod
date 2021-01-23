module github.com/gopub/wine/exp/sms

go 1.15

require (
	github.com/google/uuid v1.2.0
	github.com/gopub/conv v0.5.0 // indirect
	github.com/gopub/environ v0.3.5
	github.com/gopub/errors v0.1.7
	github.com/gopub/log v1.2.5
	github.com/gopub/types v0.3.4
	github.com/gopub/wine v1.38.0
	github.com/gopub/wine/httpvalue v0.1.4 // indirect
	golang.org/x/sys v0.0.0-20210122235752-a8b976e07c7b // indirect
)

//replace (
//	github.com/gopub/wine => ../../
//	github.com/gopub/wine/httpvalue => ../../httpvalue
//	github.com/gopub/wine/router => ../../router
//)

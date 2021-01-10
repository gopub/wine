module github.com/gopub/wine/websocket

go 1.15

require (
	github.com/golang/protobuf v1.4.3
	github.com/gopub/conv v0.3.27
	github.com/gopub/environ v0.3.5
	github.com/gopub/errors v0.1.7
	github.com/gopub/log v1.2.4
	github.com/gopub/types v0.2.25
	github.com/gopub/wine v1.36.13
	github.com/gopub/wine/router v0.1.1
	github.com/gorilla/websocket v1.4.2
	google.golang.org/protobuf v1.25.0
)

replace (
	github.com/gopub/wine => ../
	github.com/gopub/wine/httpvalue => ../httpvalue
	github.com/gopub/wine/router => ../router
)

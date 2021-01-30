module github.com/gopub/wine

go 1.15

require (
	github.com/golang/protobuf v1.4.3
	github.com/google/go-cmp v0.5.0
	github.com/google/uuid v1.2.0
	github.com/gopub/conv v0.5.1
	github.com/gopub/environ v0.3.5
	github.com/gopub/errors v0.1.7
	github.com/gopub/log v1.2.5
	github.com/gopub/types v0.3.4
	github.com/gopub/wine/httpvalue v0.1.4
	github.com/gopub/wine/router v0.1.4
	github.com/gopub/wine/urlutil v0.1.0
	github.com/gorilla/websocket v1.4.2
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/stretchr/testify v1.6.1
	golang.org/x/sys v0.0.0-20210124154548-22da62e12c0c // indirect
	golang.org/x/text v0.3.5 // indirect
	google.golang.org/protobuf v1.25.0
)

//replace (
//	github.com/gopub/wine/httpvalue => ./httpvalue
//	github.com/gopub/wine/router => ./router
//	github.com/gopub/wine/urlutil => ./urlutil
//)

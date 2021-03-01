module github.com/gopub/wine

go 1.15

require (
	github.com/golang/geo v0.0.0-20210211234256-740aa86cb551 // indirect
	github.com/golang/protobuf v1.4.3
	github.com/google/go-cmp v0.5.0
	github.com/google/uuid v1.2.0
	github.com/gopub/conv v0.6.1
	github.com/gopub/environ v0.3.5
	github.com/gopub/errors v0.1.7
	github.com/gopub/log v1.2.8
	github.com/gopub/types v0.3.19
	github.com/gopub/wine/httpvalue v0.1.4
	github.com/gopub/wine/router v0.1.4
	github.com/gopub/wine/urlutil v0.1.5
	github.com/gorilla/websocket v1.4.2
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/nyaruka/phonenumbers v1.0.66 // indirect
	github.com/stretchr/testify v1.6.1
	golang.org/x/sys v0.0.0-20210228012217-479acdf4ea46 // indirect
	golang.org/x/text v0.3.5 // indirect
	google.golang.org/protobuf v1.25.0
)

//replace (
//	github.com/gopub/wine/httpvalue => ./httpvalue
//	github.com/gopub/wine/router => ./router
//	github.com/gopub/wine/urlutil => ./urlutil
//)

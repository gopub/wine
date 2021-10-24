module github.com/gopub/wine

go 1.16

require (
	github.com/gabriel-vasile/mimetype v1.4.0 // indirect
	github.com/golang/geo v0.0.0-20210211234256-740aa86cb551 // indirect
	github.com/google/go-cmp v0.5.6
	github.com/google/uuid v1.3.0
	github.com/gopub/conv v0.6.1
	github.com/gopub/environ v0.3.5
	github.com/gopub/errors v0.1.7
	github.com/gopub/log v1.2.9
	github.com/gopub/types v0.3.19
	github.com/gopub/wine/httpvalue v0.1.4
	github.com/gopub/wine/router v0.1.5
	github.com/gopub/wine/urlutil v0.1.5
	github.com/gorilla/websocket v1.4.2
	github.com/nyaruka/phonenumbers v1.0.72 // indirect
	github.com/shopspring/decimal v1.3.0 // indirect
	github.com/spf13/viper v1.9.0 // indirect
	github.com/stretchr/testify v1.7.0
	golang.org/x/net v0.0.0-20211015210444-4f30a5c0130f // indirect
	golang.org/x/sys v0.0.0-20211015200801-69063c4bb744 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/protobuf v1.27.1
)

//replace (
//	github.com/gopub/wine/httpvalue => ./httpvalue
//	github.com/gopub/wine/router => ./router
//	github.com/gopub/wine/urlutil => ./urlutil
//)

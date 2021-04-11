module github.com/gopub/wine

go 1.16

require (
	github.com/gabriel-vasile/mimetype v1.2.0 // indirect
	github.com/golang/geo v0.0.0-20210211234256-740aa86cb551 // indirect
	github.com/golang/protobuf v1.5.2
	github.com/google/go-cmp v0.5.5
	github.com/google/uuid v1.2.0
	github.com/gopub/conv v0.6.1
	github.com/gopub/environ v0.3.5
	github.com/gopub/errors v0.1.7
	github.com/gopub/log v1.2.8
	github.com/gopub/types v0.3.19
	github.com/gopub/wine/httpvalue v0.1.4
	github.com/gopub/wine/router v0.1.5
	github.com/gopub/wine/urlutil v0.1.5
	github.com/gorilla/websocket v1.4.2
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/nyaruka/phonenumbers v1.0.68 // indirect
	github.com/pelletier/go-toml v1.9.0 // indirect
	github.com/spf13/afero v1.6.0 // indirect
	github.com/stretchr/testify v1.6.1
	golang.org/x/sys v0.0.0-20210403161142-5e06dd20ab57 // indirect
	golang.org/x/text v0.3.6 // indirect
	google.golang.org/protobuf v1.26.0
)

//replace (
//	github.com/gopub/wine/httpvalue => ./httpvalue
//	github.com/gopub/wine/router => ./router
//	github.com/gopub/wine/urlutil => ./urlutil
//)

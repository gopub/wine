module github.com/gopub/wine

go 1.15

require (
	github.com/golang/protobuf v1.4.3
	github.com/google/go-cmp v0.5.0
	github.com/google/uuid v1.1.5
	github.com/gopub/conv v0.5.0
	github.com/gopub/environ v0.3.5
	github.com/gopub/errors v0.1.7
	github.com/gopub/log v1.2.4
	github.com/gopub/types v0.3.4
	github.com/gopub/wine/httpvalue v0.1.3
	github.com/gopub/wine/router v0.1.3
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/nyaruka/phonenumbers v1.0.61 // indirect
	github.com/stretchr/testify v1.6.1
	golang.org/x/sys v0.0.0-20210119212857-b64e53b001e4 // indirect
	golang.org/x/text v0.3.5 // indirect
)

replace (
	github.com/gopub/wine/httpvalue => ./httpvalue
	github.com/gopub/wine/router => ./router
)

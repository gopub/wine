module github.com/gopub/wine/exp/vfs

go 1.15

require (
	github.com/gabriel-vasile/mimetype v1.2.0
	github.com/golang/geo v0.0.0-20210211234256-740aa86cb551 // indirect
	github.com/golang/protobuf v1.5.1 // indirect
	github.com/google/uuid v1.2.0
	github.com/gopub/conv v0.6.1
	github.com/gopub/errors v0.1.7
	github.com/gopub/log v1.2.8
	github.com/gopub/sql v1.4.17
	github.com/gopub/types v0.3.19
	github.com/gopub/wine/httpvalue v0.1.4
	github.com/mattn/go-sqlite3 v1.14.6 // indirect
	github.com/nyaruka/phonenumbers v1.0.68 // indirect
	github.com/stretchr/testify v1.6.1
)

//replace github.com/gopub/wine/httpvalue => ../../httpvalue

module github.com/gopub/wine/exp/storage/gcp

go 1.15

require (
	cloud.google.com/go v0.75.0 // indirect
	cloud.google.com/go/storage v1.12.0
	github.com/gopub/wine v1.38.2 // indirect
	github.com/gopub/wine/exp/storage v0.1.6
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/oauth2 v0.0.0-20210113205817-d3ed898aa8a3 // indirect
	google.golang.org/genproto v0.0.0-20210122163508-8081c04a3579 // indirect
	google.golang.org/grpc v1.35.0 // indirect
)

//replace (
//	github.com/gopub/wine => ../../../
//	github.com/gopub/wine/exp/storage => ../
//	github.com/gopub/wine/httpvalue => ../../../httpvalue
//	github.com/gopub/wine/router => ../../../router
//)

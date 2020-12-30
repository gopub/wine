module github.com/gopub/wine/exp/storage/gcp

go 1.15

require (
	cloud.google.com/go v0.74.0 // indirect
	cloud.google.com/go/storage v1.12.0
	github.com/gopub/log v1.2.3
	github.com/gopub/wine/exp/storage v0.1.3
	golang.org/x/net v0.0.0-20201224014010-6772e930b67b // indirect
	golang.org/x/tools v0.0.0-20201230163300-2152f4ed8ce7 // indirect
	google.golang.org/genproto v0.0.0-20201214200347-8c77b98c765d // indirect
)

replace (
	github.com/gopub/wine => ../../../
	github.com/gopub/wine/exp/storage => ../
	github.com/gopub/wine/httpvalue => ../../../httpvalue
	github.com/gopub/wine/router => ../../../router
)

module github.com/gopub/wine/exp/storage

go 1.15

require (
	github.com/disintegration/imaging v1.6.2
	github.com/google/uuid v1.2.0
	github.com/gopub/conv v0.5.0 // indirect
	github.com/gopub/errors v0.1.7
	github.com/gopub/log v1.2.5 // indirect
	github.com/gopub/wine v1.38.0
	github.com/gopub/wine/httpvalue v0.1.4
	golang.org/x/image v0.0.0-20201208152932-35266b937fa6
	golang.org/x/sys v0.0.0-20210122235752-a8b976e07c7b // indirect
)

//replace (
//	github.com/gopub/wine => ../../
//	github.com/gopub/wine/httpvalue => ../../httpvalue
//	github.com/gopub/wine/router => ../../router
//)

module github.com/gopub/wine/exp/vfs/_examples

go 1.15

require (
	github.com/gopub/environ v0.3.5 // indirect
	github.com/gopub/log v1.2.3
	github.com/gopub/wine v1.36.0
	github.com/gopub/wine/exp/vfs v0.0.0
)

replace (
	github.com/gopub/wine => ../../../
	github.com/gopub/wine/exp/vfs => ../
)

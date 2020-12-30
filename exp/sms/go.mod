module github.com/gopub/wine/exp/sms

go 1.15

require (
	github.com/gopub/environ v0.3.5
	github.com/gopub/errors v0.1.6
	github.com/gopub/log v1.2.3
	github.com/gopub/types v0.2.25
	github.com/gopub/wine v1.36.4
)

replace (
	github.com/gopub/wine => ../../
	github.com/gopub/wine/httpvalue => ../../httpvalue
	github.com/gopub/wine/router => ../../router
)

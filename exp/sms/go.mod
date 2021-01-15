module github.com/gopub/wine/exp/sms

go 1.15

require (
	github.com/google/uuid v1.1.5
	github.com/gopub/environ v0.3.5
	github.com/gopub/errors v0.1.7
	github.com/gopub/log v1.2.4
	github.com/gopub/types v0.2.27
	github.com/gopub/wine v1.37.0
)

replace (
	github.com/gopub/wine => ../../
	github.com/gopub/wine/httpvalue => ../../httpvalue
	github.com/gopub/wine/router => ../../router
)

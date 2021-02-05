module github.com/gopub/wine/session/provider

go 1.16

require (
	github.com/go-redis/redis v6.15.9+incompatible // indirect
	github.com/gopub/errors v0.1.7
	github.com/gopub/wine v1.40.2
	github.com/patrickmn/go-cache v2.1.0+incompatible
)

replace (
	github.com/gopub/wine => ../../
)
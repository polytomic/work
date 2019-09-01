module github.com/taylorchu/work/middleware/concurrent

go 1.13

replace github.com/taylorchu/work => ../../

require (
	github.com/go-redis/redis v6.15.2+incompatible
	github.com/google/uuid v1.1.1
	github.com/stretchr/testify v1.4.0
	github.com/taylorchu/work v0.0.0-00010101000000-000000000000
)

module github.com/taylorchu/work/bench

go 1.13

replace github.com/taylorchu/work => ../

require (
	github.com/go-redis/redis v6.15.2+incompatible
	github.com/gocraft/work v0.5.2-0.20180912175354-c85b71e20062
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/robfig/cron v1.2.0 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/taylorchu/work v0.0.0-00010101000000-000000000000
)

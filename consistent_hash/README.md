# Consistent Hash (*This is a community driven project*)

Adapted to Hertz's load balancing consistent-hash algorithm.

## How to use?

### Server

**[example](example/server/main.go)**

```go
package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/registry/redis"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(5)
	for i := 0; i < 5; i++ {
		num := i
		go func() {
			addr := fmt.Sprintf("127.0.0.1:800%d", num)
			r := redis.NewRedisRegistry("127.0.0.1:6379")
			h := server.Default(
				server.WithHostPorts(addr),
				server.WithRegistry(r, &registry.Info{
					ServiceName: "hertz.test.demo",
					Addr:        utils.NewNetAddr("tcp", addr),
					Weight:      10,
					Tags:        nil,
				}),
			)
			h.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
				ctx.JSON(consts.StatusOK, utils.H{"addr": addr})
			})
			h.Spin()
			wg.Done()
		}()
	}
	wg.Wait()
}
```

### Client

**[example](example/client/main.go)**

```go
package main

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/client/loadbalance"
	"github.com/cloudwego/hertz/pkg/app/middlewares/client/sd"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	lb "github.com/hertz-contrib/loadbalance/consistent_hash"
	"github.com/hertz-contrib/registry/redis"
)

func main() {
	cli, err := client.NewClient()
	if err != nil {
		hlog.Fatal(err)
		return
	}
	r := redis.NewRedisResolver("127.0.0.1:6379")
	opt := loadbalance.Options{
		RefreshInterval: 10 * time.Second,
		ExpireInterval:  25 * time.Second,
	}
	loadbalancer := lb.NewConsistentHashBalancer()
	cli.Use(sd.Discovery(r, sd.WithLoadBalanceOptions(loadbalancer, opt)))
	for i := 0; i < 10; i++ {
		status, body, err := cli.Get(context.Background(), nil, "http://hertz.test.demo/ping", config.WithSD(true))
		if err != nil {
			hlog.Fatal(err)
		}
		hlog.Infof("HERTZ: code=%d,body=%s", status, string(body))
	}
}
```

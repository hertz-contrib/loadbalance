# roundrobin (*This is a community driven project*)

Adapted to Hertz's load balancing round-robin algorithm.

## How to use?

### Server

**[example/server/main.go](server/main.go)**

```go
package main

import (
    "context"
    "fmt"
    "log"
    "sync"
    
    "github.com/cloudwego/hertz/pkg/app"
    "github.com/cloudwego/hertz/pkg/app/server"
    "github.com/cloudwego/hertz/pkg/app/server/registry"
    "github.com/cloudwego/hertz/pkg/common/utils"
    "github.com/cloudwego/hertz/pkg/protocol/consts"
    "github.com/hertz-contrib/registry/nacos"
)

func main() {
    var wg sync.WaitGroup
    wg.Add(5)
    for i := 0; i < 5; i++ {
        num := i
        go func() {
            addr := fmt.Sprintf("127.0.0.1:800%d", num)
            r, err := nacos.NewDefaultNacosRegistry()
            if err != nil {
                log.Fatal(err)
                return
            }
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

**[example/client/main.go](client/main.go)**

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/cloudwego/hertz/pkg/app/client"
    "github.com/cloudwego/hertz/pkg/app/client/loadbalance"
    "github.com/cloudwego/hertz/pkg/app/middlewares/client/sd"
    "github.com/cloudwego/hertz/pkg/common/hlog"
    "github.com/hertz-contrib/loadbalance/roundrobin"
    "github.com/hertz-contrib/registry/nacos"
)

func main() {
    cli, err := client.NewClient()
    if err != nil {
        log.Fatal(err)
        return
    }
    lb := roundrobin.NewRoundRobinBalancer()
    opt := loadbalance.Options{
        RefreshInterval: 10 * time.Second,
        ExpireInterval:  25 * time.Second,
    }
    r, err := nacos.NewDefaultNacosResolver()
    if err != nil {
        log.Fatal(err)
        return
    }
    cli.Use(sd.Discovery(r, sd.WithLoadBalanceOptions(lb, opt)))
    for i := 0; i < 10; i++ {
        status, body, err := client.Get(context.Background(), nil, "http://hertz.test.demo/ping")
        if err != nil {
            hlog.Fatal(err)
        }
        hlog.Infof("code=%d,body=%s\n", status, string(body))
    }
}
```
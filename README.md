# loadbalance (This is a community driven project)

The loadbalance extensions for Hertz, which currently implements round-robin and weighted round-robin algorithms.

## Install

``` shell
go get github.com/hertz-contrib/loadbalance
```

## import

```go
import "github.com/hertz-contrib/loadbalance"
```

## Example

```go
package main

import (
	"log"

	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/client/loadbalance"
	"github.com/cloudwego/hertz/pkg/app/middlewares/client/sd"
	loadbalanceEx "github.com/hertz-contrib/loadbalance"
	"github.com/hertz-contrib/registry/nacos"
)

func main() {
	cli, err := client.NewClient()
	r, err := nacos.NewDefaultNacosResolver()
	// ...
	lb := loadbalanceEx.NewRoundRobinBalancer()
	opt := loadbalance.Options{
		// ...
	}
	cli.Use(sd.Discovery(r, sd.WithLoadBalanceOptions(lb, opt)))
}
```

## Usage

| usage                                                      | description                                                  |
|------------------------------------------------------------|--------------------------------------------------------------|
| [round-robin](example/round_robin/main.go)                 | How to use round-robin Algorithms in Load Balancing          |
| [weighted round-robin](example/weight_round_robin/main.go) | How to use weighted round-robin Algorithms in Load Balancing |

## License

This project is under the Apache License 2.0. See the LICENSE file for the full license text.
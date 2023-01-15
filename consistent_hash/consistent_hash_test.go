/*
 * Copyright 2022 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package consistent_hash

import (
	"strconv"
	"testing"

	"github.com/cloudwego/hertz/pkg/app/client/discovery"
	"github.com/cloudwego/hertz/pkg/common/test/assert"
)

func TestConsistentHashBalancer(t *testing.T) {
	balancer := NewConsistentHashBalancer()
	// nil
	ins := balancer.Pick(discovery.Result{})
	assert.DeepEqual(t, nil, ins)

	// empty instance
	res := discovery.Result{
		CacheKey:  "a",
		Instances: make([]discovery.Instance, 0),
	}
	balancer.Rebalance(res)
	ins = balancer.Pick(res)
	assert.DeepEqual(t, nil, ins)

	// one instance
	insList := []discovery.Instance{
		discovery.NewInstance("tcp", "127.0.0.1:8888", 10, nil),
	}
	res = discovery.Result{
		CacheKey:  "b",
		Instances: insList,
	}
	balancer.Rebalance(res)
	for i := 0; i < 100; i++ {
		ins = balancer.Pick(res)
		assert.DeepEqual(t, "127.0.0.1:8888", ins.Address().String())
	}
}

func TestConsistentHashBalancerWithMulti(t *testing.T) {
	// multi instances
	balancer := NewConsistentHashBalancer(
		WithHashFunc(func(data []byte) uint32 {
			s := string(data)
			n, _ := strconv.Atoi(s[len(s)-2:])
			return uint32(n)
		}),
		WithReplicationFactor(3),
	)
	// hash("127.0.0.1:8001") => [10 11 12]
	// hash("127.0.0.1:8002") => [20 21 22]
	// hash("127.0.0.1:8003") => [30 31 32]
	// ring:
	// [10 11 12 20 21 22 30 31 32]
	// nodes:
	// [10 11 12] => 8001
	// [20 21 22] => 8002
	// [30 31 32] => 8003
	insList := []discovery.Instance{
		discovery.NewInstance("tcp", "127.0.0.1:8001", 10, nil),
		discovery.NewInstance("tcp", "127.0.0.1:8002", 10, nil),
		discovery.NewInstance("tcp", "127.0.0.1:8003", 10, nil),
	}
	res := discovery.Result{
		// hash("01") => 1
		CacheKey:  "01",
		Instances: insList,
	}
	balancer.Rebalance(res)
	ins := balancer.Pick(res)
	assert.DeepEqual(t, "127.0.0.1:8001", ins.Address().String())

	insListModified := []discovery.Instance{
		discovery.NewInstance("tcp", "127.0.0.1:8001", 10, nil),
		discovery.NewInstance("tcp", "127.0.0.1:8002", 10, nil),
		discovery.NewInstance("tcp", "127.0.0.1:8003", 10, nil),
		discovery.NewInstance("tcp", "127.0.0.1:8004", 10, nil),
		discovery.NewInstance("tcp", "127.0.0.1:8005", 10, nil),
	}
	res2 := discovery.Result{
		CacheKey:  "01",
		Instances: insListModified,
	}
	balancer.Rebalance(res2)
	ins2 := balancer.Pick(res2)
	assert.DeepEqual(t, "127.0.0.1:8001", ins2.Address().String())
}

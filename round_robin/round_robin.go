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

package roundrobin

import (
	"sync"
	"sync/atomic"

	"github.com/cloudwego/hertz/pkg/app/client/discovery"
	"github.com/cloudwego/hertz/pkg/app/client/loadbalance"
	"golang.org/x/sync/singleflight"
)

type roundRobinBalancer struct {
	cachedInfo sync.Map
	sfg        singleflight.Group
}

type roundRobinInfo struct {
	instances []discovery.Instance
	index     uint32
}

// NewRoundRobinBalancer creates a loadbalancer using round-robin algorithm.
func NewRoundRobinBalancer() loadbalance.Loadbalancer {
	lb := &roundRobinBalancer{}
	return lb
}

// Pick implements the Loadbalancer interface.
func (rr *roundRobinBalancer) Pick(e discovery.Result) discovery.Instance {
	ri, ok := rr.cachedInfo.Load(e.CacheKey)
	if !ok {
		ri, _, _ = rr.sfg.Do(e.CacheKey, func() (interface{}, error) {
			return &roundRobinInfo{
				instances: e.Instances,
				index:     0,
			}, nil
		})
		rr.cachedInfo.Store(e.CacheKey, ri)
	}

	r := ri.(*roundRobinInfo)
	if len(r.instances) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	var instance discovery.Instance
	var newIdx uint32
	wg.Add(1)
	go func() {
		newIdx = atomic.AddUint32(&r.index, 1)
		instance = r.instances[newIdx-1]
		lens := len(r.instances)
		atomic.StoreUint32(&r.index, (newIdx)%uint32(lens))
		wg.Done()
	}()
	wg.Wait()

	return instance
}

// Rebalance implements the Loadbalancer interface.
func (rr *roundRobinBalancer) Rebalance(e discovery.Result) {
	rr.cachedInfo.Store(e.CacheKey, &roundRobinInfo{
		instances: e.Instances,
		index:     0,
	})
}

// Delete implements the Loadbalancer interface.
func (rr *roundRobinBalancer) Delete(cacheKey string) {
	rr.cachedInfo.Delete(cacheKey)
}

// Name implements the Loadbalancer interface.
func (rr *roundRobinBalancer) Name() string {
	return "round_robin"
}

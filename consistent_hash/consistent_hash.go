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
	"fmt"
	"sort"
	"sync"

	"github.com/cloudwego/hertz/pkg/app/client/discovery"
	"github.com/cloudwego/hertz/pkg/app/client/loadbalance"
	"golang.org/x/sync/singleflight"
)

type consistentHashBalancer struct {
	cachedHashInfo sync.Map
	sfg            singleflight.Group
	options        *Options
	sync.RWMutex
}

type (
	consistentHashInfo struct {
		ring  Ring
		nodes map[uint32]discovery.Instance
	}
	HashFunc func(data []byte) uint32
	Ring     []uint32
)

// NewConsistentHashBalancer creates a loadbalancer using consistent-hash algorithm
func NewConsistentHashBalancer(opts ...Option) loadbalance.Loadbalancer {
	options := newOptions(opts...)
	return &consistentHashBalancer{
		options: options,
	}
}

func (ch *consistentHashBalancer) buildRing(res discovery.Result) *consistentHashInfo {
	chInfo := &consistentHashInfo{
		ring:  make([]uint32, 0),
		nodes: make(map[uint32]discovery.Instance, len(res.Instances)*ch.options.ReplicationFactor),
	}
	for _, instance := range res.Instances {
		for i := 0; i < ch.options.ReplicationFactor; i++ {
			hash := ch.options.HashFunc([]byte(fmt.Sprintf("%s%d", instance.Address().String(), i)))
			chInfo.nodes[hash] = instance
			chInfo.ring = append(chInfo.ring, hash)
		}
	}
	sort.Slice(chInfo.ring, func(i, j int) bool {
		return chInfo.ring[i] < chInfo.ring[j]
	})
	return chInfo
}

// Pick implements the loadbalance.Loadbalancer interface
func (ch *consistentHashBalancer) Pick(res discovery.Result) discovery.Instance {
	chInfo, ok := ch.cachedHashInfo.Load(res.CacheKey)
	if !ok {
		chInfo, _, _ = ch.sfg.Do(res.CacheKey, func() (any, error) {
			return ch.buildRing(res), nil
		})
		ch.cachedHashInfo.Store(res.CacheKey, chInfo)
	}
	ch.RLock()
	defer ch.RUnlock()
	info := chInfo.(*consistentHashInfo)
	if len(info.nodes) == 0 {
		return nil
	}
	hash := ch.options.HashFunc([]byte(res.CacheKey))
	idx := sort.Search(len(info.ring), func(i int) bool {
		return info.ring[i] >= hash
	})
	if idx >= len(info.ring) {
		idx = 0
	}
	return info.nodes[info.ring[idx]]
}

// Rebalance implements the loadbalance.Loadbalancer interface
func (ch *consistentHashBalancer) Rebalance(res discovery.Result) {
	ch.cachedHashInfo.Store(res.CacheKey, ch.buildRing(res))
}

// Delete implements the loadbalance.Loadbalancer interface
func (ch *consistentHashBalancer) Delete(cacheKey string) {
	ch.cachedHashInfo.Delete(cacheKey)
}

// Name implements the loadbalance.Loadbalancer interface
func (ch *consistentHashBalancer) Name() string {
	return "consistent_hash"
}

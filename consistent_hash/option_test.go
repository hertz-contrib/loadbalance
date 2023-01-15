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
	"hash/crc32"
	"strconv"
	"testing"

	"github.com/cloudwego/hertz/pkg/common/test/assert"
)

func TestOptions(t *testing.T) {
	hashFunc := func(data []byte) uint32 {
		n, _ := strconv.Atoi(string(data))
		return uint32(n)
	}
	opts := newOptions(
		WithHashFunc(hashFunc),
		WithReplicationFactor(13),
	)
	assert.DeepEqual(t, fmt.Sprintf("%p", hashFunc), fmt.Sprintf("%p", opts.HashFunc))
	assert.DeepEqual(t, 13, opts.ReplicationFactor)
}

func TestDefaultOptions(t *testing.T) {
	opts := newOptions()
	assert.DeepEqual(t, fmt.Sprintf("%p", crc32.ChecksumIEEE), fmt.Sprintf("%p", opts.HashFunc))
	assert.DeepEqual(t, 10, opts.ReplicationFactor)
}

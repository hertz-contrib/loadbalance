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

import "hash/crc32"

type Option struct {
	F func(o *Options)
}

type Options struct {
	HashFunc          HashFunc
	ReplicationFactor int
}

func newOptions(opts ...Option) *Options {
	options := &Options{
		HashFunc:          crc32.ChecksumIEEE,
		ReplicationFactor: 10,
	}
	options.apply(opts...)
	return options
}

func (o *Options) apply(opts ...Option) {
	for _, opt := range opts {
		opt.F(o)
	}
}

func WithHashFunc(hashFunc HashFunc) Option {
	return Option{F: func(o *Options) {
		o.HashFunc = hashFunc
	}}
}

func WithReplicationFactor(num int) Option {
	return Option{F: func(o *Options) {
		o.ReplicationFactor = num
	}}
}

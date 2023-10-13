// Copyright 2023 ecodeclub
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ratelimit

import "context"

//go:generate mockgen -source=types.go -package=limitmocks -destination=./mocks/ratelimit.mock.go
type Limiter interface {
	// Limit 有没有触发限流。key 就是限流对象
	// bool 代表是否限流，true 就是要限流
	// err 限流器本身有没有错误
	Limit(ctx context.Context, key string) (bool, error)
}

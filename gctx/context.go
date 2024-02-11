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

package gctx

import (
	"github.com/ecodeclub/ekit"
	"github.com/gin-gonic/gin"
)

type Context struct {
	*gin.Context
}

func (c *Context) Param(key string) ekit.AnyValue {
	return ekit.AnyValue{
		Val: c.Context.Param(key),
	}
}

func (c *Context) Query(key string) ekit.AnyValue {
	return ekit.AnyValue{
		Val: c.Context.Query(key),
	}
}

func (c *Context) Cookie(key string) ekit.AnyValue {
	val, err := c.Context.Cookie(key)
	return ekit.AnyValue{
		Val: val,
		Err: err,
	}
}

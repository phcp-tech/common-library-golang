// Copyright(C) 2019-2026 PHCP Technologies. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// 	http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test_test

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	libtest "github.com/phcp-tech/common-library-golang/test"
)

// TestNewHttpExpect is the regression test backing ExampleNewHttpExpect (see
// example_test.go): it drives the same GET/POST/error-path requests but with
// a real *testing.T, so a broken NewHttpExpect wiring actually fails the
// build instead of only being caught downstream by consumer repos.
func TestNewHttpExpect(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	router.POST("/echo", func(c *gin.Context) {
		var body map[string]any
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": body})
	})

	router.GET("/secret", func(c *gin.Context) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	})

	e := libtest.NewHttpExpect(t, router)

	// GET — check status code and a JSON string field.
	e.GET("/ping").
		Expect().
		Status(http.StatusOK).
		JSON().Object().
		Value("message").String().IsEqual("pong")

	// POST — send a JSON body; assert a nested response field.
	e.POST("/echo").
		WithJSON(map[string]any{"name": "world"}).
		Expect().
		Status(http.StatusOK).
		JSON().Object().
		Value("data").Object().
		Value("name").String().IsEqual("world")

	// Error path — assert a non-2xx status code.
	e.GET("/secret").
		Expect().
		Status(http.StatusUnauthorized).
		JSON().Object().
		Value("error").String().IsEqual("unauthorized")
}

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

package network_test

import (
	"fmt"
	"net/http"

	"github.com/phcp-tech/common-library-golang/network"
)

// ExampleGetRemoteIp shows how to extract the real client IP in a reverse-proxy
// environment where the original IP is forwarded via X-Forwarded-For.
func ExampleGetRemoteIp() {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.5")
	req.RemoteAddr = "10.0.0.1:1234"

	fmt.Println(network.GetRemoteIp(req))
	// Output:
	// 203.0.113.5
}

// ExampleGetRemoteIp_multipleProxies shows that when X-Forwarded-For contains
// multiple addresses (set by a chain of proxies), the first (original client)
// IP is returned.
func ExampleGetRemoteIp_multipleProxies() {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.5, 10.0.0.2, 10.0.0.3")
	req.RemoteAddr = "10.0.0.1:1234"

	fmt.Println(network.GetRemoteIp(req))
	// Output:
	// 203.0.113.5
}

// ExampleGetRemoteIp_xRealIP shows the X-Real-IP fallback used by Nginx when
// X-Forwarded-For is not set.
func ExampleGetRemoteIp_xRealIP() {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-IP", "198.51.100.7")
	req.RemoteAddr = "10.0.0.1:1234"

	fmt.Println(network.GetRemoteIp(req))
	// Output:
	// 198.51.100.7
}

// ExampleGetLocalIpAddress shows how to retrieve all non-loopback IPv4 addresses
// assigned to the local machine — useful for logging the pod IP at startup.
func ExampleGetLocalIpAddress() {
	addrs := network.GetLocalIpAddress()
	for _, ip := range addrs {
		fmt.Println(ip)
	}
}

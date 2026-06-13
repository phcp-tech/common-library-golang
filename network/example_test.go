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

// ExampleInt2IpWithLittleEndian converts a little-endian uint32 to a dotted-decimal
// IPv4 string. This format is used by MT4/MT5 trading platforms.
func ExampleInt2IpWithLittleEndian() {
	fmt.Println(network.Int2IpWithLittleEndian(0x04030201)) // 0x04030201 = 1.2.3.4
	fmt.Println(network.Int2IpWithLittleEndian(0))          // 0 → empty
	// Output:
	// 1.2.3.4
	//
}

// ExampleInt2IpWithBigEndian converts a big-endian uint32 (network byte order)
// to a dotted-decimal IPv4 string.
func ExampleInt2IpWithBigEndian() {
	fmt.Println(network.Int2IpWithBigEndian(0x01020304)) // 0x01020304 = 1.2.3.4
	fmt.Println(network.Int2IpWithBigEndian(0x7F000001)) // 0x7F000001 = 127.0.0.1
	// Output:
	// 1.2.3.4
	// 127.0.0.1
}

// ExampleIp2IntWithLittleEndian converts a dotted-decimal IPv4 string to a
// little-endian uint32. Returns 0 for empty or invalid input.
func ExampleIp2IntWithLittleEndian() {
	fmt.Printf("0x%X\n", network.Ip2IntWithLittleEndian("1.2.3.4"))
	fmt.Println(network.Ip2IntWithLittleEndian(""))
	// Output:
	// 0x4030201
	// 0
}

// ExampleIp2IntWithBigEndian converts a dotted-decimal IPv4 string to a
// big-endian uint32 (network byte order). Returns 0 for empty or invalid input.
func ExampleIp2IntWithBigEndian() {
	fmt.Printf("0x%X\n", network.Ip2IntWithBigEndian("1.2.3.4"))
	fmt.Printf("0x%X\n", network.Ip2IntWithBigEndian("127.0.0.1"))
	// Output:
	// 0x1020304
	// 0x7F000001
}

// ExampleIsValidAddr checks whether a string is a valid IP address or
// resolvable hostname, with or without a port.
func ExampleIsValidAddr() {
	fmt.Println(network.IsValidAddr("192.168.1.1:8080")) // IP with port
	fmt.Println(network.IsValidAddr("192.168.1.1"))      // IP without port
	fmt.Println(network.IsValidAddr("999.999.999.999"))  // invalid
	// Output:
	// true
	// true
	// false
}

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

package app

import (
	"net"
	"net/http"
	"strings"
)

// GetRemoteIp returns the remote client IP address from the HTTP request.
// It inspects the X-Forwarded-For header first (required in AWS Lambda and other
// reverse-proxy environments), then X-Real-IP, and finally falls back to the
// TCP remote address. The loopback address "::1" is normalised to "127.0.0.1".
// When X-Forwarded-For contains multiple addresses, the first one is returned.
func GetRemoteIp(req *http.Request) string {
	remoteAddr := req.RemoteAddr
	if ip := req.Header.Get("X-Forwarded-For"); ip != "" {
		remoteAddr = ip
	} else if ip = req.Header.Get("X-Real-IP"); ip != "" {
		remoteAddr = ip
	} else {
		remoteAddr, _, _ = net.SplitHostPort(remoteAddr)
	}

	if remoteAddr == "::1" {
		remoteAddr = "127.0.0.1"
	}

	// return the first IP address
	if addrList := strings.Split(remoteAddr, ","); len(addrList) > 0 {
		return addrList[0]
	} else {
		return ""
	}
}

// GetLocalIpAddress returns all non-loopback IPv4 addresses assigned to the
// local machine's network interfaces.
func GetLocalIpAddress() (ipaddr []string) {
	if addrs, err := net.InterfaceAddrs(); err == nil {
		for _, address := range addrs {
			// except 127.0.0.1
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					ipaddr = append(ipaddr, ipnet.IP.String())
				}
			}
		}
	}
	return ipaddr
}

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

package network

import (
	"net"
	"net/http"
	"testing"
)

// -----------------------------------------------------------------------
// GetRemoteIp
// -----------------------------------------------------------------------

func TestGetRemoteIp(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		xForwarded string
		xRealIP    string
		want       string
	}{
		{
			name:       "X-Forwarded-For single IP",
			remoteAddr: "10.0.0.1:1234",
			xForwarded: "203.0.113.5",
			want:       "203.0.113.5",
		},
		{
			name:       "X-Forwarded-For multiple IPs returns first",
			remoteAddr: "10.0.0.1:1234",
			xForwarded: "203.0.113.5, 10.0.0.2, 10.0.0.3",
			want:       "203.0.113.5",
		},
		{
			name:       "X-Real-IP used when no X-Forwarded-For",
			remoteAddr: "10.0.0.1:1234",
			xRealIP:    "198.51.100.7",
			want:       "198.51.100.7",
		},
		{
			name:       "X-Forwarded-For takes precedence over X-Real-IP",
			remoteAddr: "10.0.0.1:1234",
			xForwarded: "203.0.113.5",
			xRealIP:    "198.51.100.7",
			want:       "203.0.113.5",
		},
		{
			name:       "fallback to RemoteAddr host:port",
			remoteAddr: "192.168.1.100:5678",
			want:       "192.168.1.100",
		},
		{
			name:       "IPv6 loopback normalised to 127.0.0.1",
			remoteAddr: "[::1]:1234",
			want:       "127.0.0.1",
		},
		{
			name:       "X-Forwarded-For IPv6 loopback normalised",
			remoteAddr: "10.0.0.1:1234",
			xForwarded: "::1",
			want:       "127.0.0.1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tc.remoteAddr
			if tc.xForwarded != "" {
				req.Header.Set("X-Forwarded-For", tc.xForwarded)
			}
			if tc.xRealIP != "" {
				req.Header.Set("X-Real-IP", tc.xRealIP)
			}

			got := GetRemoteIp(req)
			if got != tc.want {
				t.Errorf("GetRemoteIp() = %q, want %q", got, tc.want)
			}
		})
	}
}

// -----------------------------------------------------------------------
// GetLocalIpAddress
// -----------------------------------------------------------------------

func TestGetLocalIpAddress_ReturnsValidIPv4(t *testing.T) {
	addrs := GetLocalIpAddress()
	// On any machine with a network interface this should be non-nil;
	// on a CI runner with only loopback it may be empty — both are valid.
	for _, addr := range addrs {
		ip := net.ParseIP(addr)
		if ip == nil {
			t.Errorf("GetLocalIpAddress() returned invalid IP %q", addr)
			continue
		}
		if ip.To4() == nil {
			t.Errorf("GetLocalIpAddress() returned non-IPv4 address %q", addr)
		}
		if ip.IsLoopback() {
			t.Errorf("GetLocalIpAddress() returned loopback address %q", addr)
		}
	}
}

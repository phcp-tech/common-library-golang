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
// Int2IpWithLittleEndian / Int2IpWithBigEndian
// -----------------------------------------------------------------------

func TestInt2IpWithLittleEndian(t *testing.T) {
	tests := []struct {
		name  string
		input uint32
		want  string
	}{
		{"zero returns empty", 0, ""},
		{"1.2.3.4 little-endian", 0x04030201, "1.2.3.4"},
		{"10.0.0.1 little-endian", 0x0100000A, "10.0.0.1"},
		{"127.0.0.1 little-endian", 0x0100007F, "127.0.0.1"},
		{"255.255.255.255 little-endian", 0xFFFFFFFF, "255.255.255.255"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Int2IpWithLittleEndian(tc.input)
			if got != tc.want {
				t.Errorf("Int2IpWithLittleEndian(%d) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestInt2IpWithBigEndian(t *testing.T) {
	tests := []struct {
		name  string
		input uint32
		want  string
	}{
		{"zero returns empty", 0, ""},
		{"1.2.3.4 big-endian", 0x01020304, "1.2.3.4"},
		{"10.0.0.1 big-endian", 0x0A000001, "10.0.0.1"},
		{"127.0.0.1 big-endian", 0x7F000001, "127.0.0.1"},
		{"255.255.255.255 big-endian", 0xFFFFFFFF, "255.255.255.255"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Int2IpWithBigEndian(tc.input)
			if got != tc.want {
				t.Errorf("Int2IpWithBigEndian(%d) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// -----------------------------------------------------------------------
// Ip2IntWithLittleEndian / Ip2IntWithBigEndian
// -----------------------------------------------------------------------

func TestIp2IntWithLittleEndian(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  uint32
	}{
		{"empty returns 0", "", 0},
		{"invalid returns 0", "not-an-ip", 0},
		{"IPv6 returns 0", "::1", 0},
		{"1.2.3.4", "1.2.3.4", 0x04030201},
		{"10.0.0.1", "10.0.0.1", 0x0100000A},
		{"127.0.0.1", "127.0.0.1", 0x0100007F},
		{"255.255.255.255", "255.255.255.255", 0xFFFFFFFF},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Ip2IntWithLittleEndian(tc.input)
			if got != tc.want {
				t.Errorf("Ip2IntWithLittleEndian(%q) = %d (0x%X), want %d (0x%X)", tc.input, got, got, tc.want, tc.want)
			}
		})
	}
}

func TestIp2IntWithBigEndian(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  uint32
	}{
		{"empty returns 0", "", 0},
		{"invalid returns 0", "not-an-ip", 0},
		{"IPv6 returns 0", "::1", 0},
		{"1.2.3.4", "1.2.3.4", 0x01020304},
		{"10.0.0.1", "10.0.0.1", 0x0A000001},
		{"127.0.0.1", "127.0.0.1", 0x7F000001},
		{"255.255.255.255", "255.255.255.255", 0xFFFFFFFF},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Ip2IntWithBigEndian(tc.input)
			if got != tc.want {
				t.Errorf("Ip2IntWithBigEndian(%q) = %d (0x%X), want %d (0x%X)", tc.input, got, got, tc.want, tc.want)
			}
		})
	}
}

// -----------------------------------------------------------------------
// Roundtrip: Ip2Int → Int2Ip
// -----------------------------------------------------------------------

func TestRoundtrip_LittleEndian(t *testing.T) {
	ips := []string{"1.2.3.4", "10.0.0.1", "192.168.1.100", "255.255.255.255"}
	for _, ip := range ips {
		n := Ip2IntWithLittleEndian(ip)
		got := Int2IpWithLittleEndian(n)
		if got != ip {
			t.Errorf("little-endian roundtrip %q → %d → %q", ip, n, got)
		}
	}
}

func TestRoundtrip_BigEndian(t *testing.T) {
	ips := []string{"1.2.3.4", "10.0.0.1", "192.168.1.100", "255.255.255.255"}
	for _, ip := range ips {
		n := Ip2IntWithBigEndian(ip)
		got := Int2IpWithBigEndian(n)
		if got != ip {
			t.Errorf("big-endian roundtrip %q → %d → %q", ip, n, got)
		}
	}
}

// -----------------------------------------------------------------------
// IsValidAddr
// -----------------------------------------------------------------------

func TestIsValidAddr(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid IP with port", "192.168.1.1:8080", true},
		{"valid IP without port", "192.168.1.1", true},
		{"loopback with port", "127.0.0.1:3306", true},
		{"loopback without port", "127.0.0.1", true},
		{"invalid IP", "999.999.999.999", false},
		{"empty string", "", false},
		{"port only", ":8080", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := IsValidAddr(tc.input)
			if got != tc.want {
				t.Errorf("IsValidAddr(%q) = %v, want %v", tc.input, got, tc.want)
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

func TestIsPrivateIP(t *testing.T) {
	cases := []struct {
		name string
		ip   string
		want bool
	}{
		// RFC 1918 private IPv4 ranges: should be private
		{"10.0.0.0/8 lower bound", "10.0.0.0", true},
		{"10.0.0.0/8 typical", "10.0.0.1", true},
		{"10.0.0.0/8 upper bound", "10.255.255.255", true},
		{"172.16.0.0/12 lower bound", "172.16.0.0", true},
		{"172.16.0.0/12 typical", "172.16.5.5", true},
		{"172.16.0.0/12 upper bound", "172.31.255.255", true},
		{"192.168.0.0/16 lower bound", "192.168.0.0", true},
		{"192.168.0.0/16 typical", "192.168.1.5", true},
		{"192.168.0.0/16 upper bound", "192.168.255.255", true},

		// just outside RFC 1918 boundaries: should NOT be private
		{"just below 172.16.0.0/12", "172.15.255.255", false},
		{"just above 172.16.0.0/12", "172.32.0.0", false},
		{"just below 192.168.0.0/16", "192.167.255.255", false},

		// 169.254.0.0/16 link-local: should NOT be private
		{"169.254.0.0/16 lower bound", "169.254.0.0", false},
		{"169.254.0.0/16 typical", "169.254.0.1", false},
		{"169.254.0.0/16 address from conversation", "169.254.0.2", false},
		{"169.254.0.0/16 mid range", "169.254.128.128", false},
		{"169.254.0.0/16 upper bound", "169.254.255.255", false},
		{"just below 169.254.0.0/16", "169.253.255.255", false},
		{"just above 169.254.0.0/16", "169.255.0.0", false},

		// public IPv4 addresses: should NOT be private
		{"public IP: google dns", "8.8.8.8", false},
		{"public IP: cloudflare dns", "1.1.1.1", false},

		// loopback: should NOT be private
		{"loopback", "127.0.0.1", false},

		// RFC 4193 IPv6 unique local addresses: should be private
		{"IPv6 unique local fc00::/7", "fc00::1", true},
		{"IPv6 unique local fd00::/8", "fd12:3456:789a::1", true},

		// IPv6 link-local: should NOT be private
		{"IPv6 link-local fe80::/10", "fe80::1", false},

		// invalid or empty input: should NOT be private
		{"empty string", "", false},
		{"not an IP", "not-an-ip", false},
		{"out-of-range octets", "999.999.999.999", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsPrivateIP(tc.ip); got != tc.want {
				t.Errorf("IsPrivateIP(%q) = %v, want %v", tc.ip, got, tc.want)
			}
		})
	}
}

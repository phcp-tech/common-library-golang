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
	"encoding/binary"
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

// Int2IpWithLittleEndian converts a uint32 IP address to string format in little-endian format
func Int2IpWithLittleEndian(ipInt uint32) string {
	// Handle special case when IP is 0
	if ipInt == 0 {
		return ""
	}

	// Convert uint32 to 4 bytes (IPv4)
	// MT4 stores IP addresses in little-endian byte order
	byte1 := ipInt & 0xFF
	byte2 := (ipInt >> 8) & 0xFF
	byte3 := (ipInt >> 16) & 0xFF
	byte4 := (ipInt >> 24) & 0xFF

	// Create IPv4 address and convert to string
	ip := net.IPv4(byte(byte1), byte(byte2), byte(byte3), byte(byte4))
	return ip.String()
}

// Int2IpWithBigEndian converts a uint32 IP address to string format in big-endian format
func Int2IpWithBigEndian(ipInt uint32) string {
	// Handle special case when IP is 0
	if ipInt == 0 {
		return ""
	}

	// Convert uint32 to 4 bytes (IPv4)
	// Big-endian byte order - most significant byte first
	byte1 := (ipInt >> 24) & 0xFF
	byte2 := (ipInt >> 16) & 0xFF
	byte3 := (ipInt >> 8) & 0xFF
	byte4 := ipInt & 0xFF

	// Create IPv4 address and convert to string
	ip := net.IPv4(byte(byte1), byte(byte2), byte(byte3), byte(byte4))
	return ip.String()
}

// Ip2IntWithLittleEndian converts an IP address string to uint32 in little-endian format
func Ip2IntWithLittleEndian(ipStr string) uint32 {
	// Handle special case when IP string is empty
	if ipStr == "" {
		return 0
	}

	// Parse IP string to net.IP
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return 0
	}

	// Convert to IPv4 (4 bytes)
	ip = ip.To4()
	if ip == nil {
		return 0
	}

	// Convert to uint32 using little-endian byte order
	// Little-endian: least significant byte first
	//return uint32(ip[0]) | (uint32(ip[1]) << 8) | (uint32(ip[2]) << 16) | (uint32(ip[3]) << 24)
	return binary.LittleEndian.Uint32(ip)
}

// Ip2IntWithBigEndian converts an IP address string to uint32 in big-endian format
func Ip2IntWithBigEndian(ipStr string) uint32 {
	// Handle special case when IP string is empty
	if ipStr == "" {
		return 0
	}

	// Parse IP string to net.IP
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return 0
	}

	// Convert to IPv4 (4 bytes)
	ip = ip.To4()
	if ip == nil {
		return 0
	}

	// Convert to uint32 using big-endian byte order
	// Big-endian: most significant byte first
	//return (uint32(ip[0]) << 24) | (uint32(ip[1]) << 16) | (uint32(ip[2]) << 8) | uint32(ip[3])
	return binary.BigEndian.Uint32(ip)
}

// IsPrivateIP reports whether ipStr is a private IP address, i.e. within
// RFC 1918 (IPv4: 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16) or RFC 4193
// (IPv6 unique local addresses: fc00::/7). Link-local addresses (e.g.
// 169.254.0.0/16) are NOT private and return false. An unparsable ipStr
// also returns false.
func IsPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	return ip.IsPrivate()
}

// IsValidAddr checks if the given string is a valid IP address with port or resolvable domain name with port
func IsValidAddr(ipStr string) bool {
	// split host and port
	host, _, err := net.SplitHostPort(ipStr)
	if err != nil {
		// if there is no port, use the whole string as host
		host = ipStr
	}

	// Check if it's a valid IP address
	ip := net.ParseIP(host)
	if ip == nil {
		// Not an IP, try DNS lookup to validate domain name
		_, err := net.LookupHost(host)
		if err != nil {
			return false
		}
		return true
	}

	return true
}

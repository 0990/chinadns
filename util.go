package chinadns

import (
	"net"
	"time"
)

type IPType int

const (
	IPInvalid IPType = iota
	IPV4
	IPV6
)

func timeSinceMS(t time.Time) int64 {
	return int64(time.Since(t) / 1e6)
}

func getIPType(ip string) IPType {
	if net.ParseIP(ip) == nil {
		return IPInvalid
	}
	for i := 0; i < len(ip); i++ {
		switch ip[i] {
		case '.':
			return IPV4
		case ':':
			return IPV6
		}
	}

	return IPInvalid
}

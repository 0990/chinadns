package chinadns

import (
	"github.com/miekg/dns"
	"net"
	"strings"
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

func getIPV4(vs []string) (ret []string) {
	for _, v := range vs {
		if getIPType(v) == IPV4 {
			ret = append(ret, v)
		}
	}
	return
}

func getIPV6(vs []string) (ret []string) {
	for _, v := range vs {
		if getIPType(v) == IPV6 {
			ret = append(ret, v)
		}
	}
	return
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

func IsInArray[T comparable](elem T, arr []T) bool {
	return FindIndex(arr, elem) >= 0
}

func FindIndex[T comparable](arr []T, elem T) int {
	for i, v := range arr {
		if v == elem {
			return i
		}
	}
	return -1
}

func IsContains(s string, substrs []string) bool {
	for _, v := range substrs {
		if strings.Contains(s, v) {
			return true
		}
	}

	return false
}

func replyIP(reply *dns.Msg) []net.IP {
	var ip []net.IP
	for _, rr := range reply.Answer {
		switch answer := rr.(type) {
		case *dns.A:
			ip = append(ip, answer.A)
		case *dns.AAAA:
			ip = append(ip, answer.AAAA)
		case *dns.CNAME:
			continue
		default:
			continue
		}
	}
	return ip
}

func replyCDName(reply *dns.Msg) (ret []string) {
	for _, rr := range reply.Answer {
		switch answer := rr.(type) {
		case *dns.A:
			continue
		case *dns.AAAA:
			continue
		case *dns.CNAME:
			ret = append(ret, answer.Target)
		case *dns.DNAME:
			ret = append(ret, answer.Target)
		default:
			continue
		}
	}
	return ret
}

func answerIPString(reply *dns.Msg) string {
	ips := replyIP(reply)
	var ip string
	for _, v := range ips {
		ip += v.String()
		ip += ";"
	}
	return ip
}

func answerCDNameString(reply *dns.Msg) string {
	ips := replyCDName(reply)
	return strings.Join(ips, ";")
}

func replyString(reply *dns.Msg) string {
	if reply == nil {
		return ""
	}
	return answerIPString(reply) + answerCDNameString(reply)
}

package chinadns

import (
	"github.com/miekg/dns"
)

var DefaultDNSAdBlockReply = []string{"0.0.0.0", "::", "fake-for-negative-caching.adguard.com"}

type AdBlockJudge struct {
	ipv4s   []string
	ipv6s   []string
	domains []string
}

func NewAdBlockJudge(reply []string) *AdBlockJudge {
	j := &AdBlockJudge{}

	if len(reply) == 0 {
		reply = DefaultDNSAdBlockReply
	}

	for _, v := range reply {
		typ := getIPType(v)
		switch typ {
		case IPV4:
			j.ipv4s = append(j.ipv4s, v)
		case IPV6:
			j.ipv6s = append(j.ipv6s, v)
		default:
			j.domains = append(j.domains, v)
		}
	}
	return j
}

func (j *AdBlockJudge) IsAdBlockReply(reply *dns.Msg) bool {
	if reply == nil {
		return false
	}

	if len(reply.Answer) > 0 {
		answer := reply.Answer[0]
		switch answer.Header().Rrtype {
		case dns.TypeA:
			s := answer.(*dns.A).A.String()
			return IsInArray(s, j.ipv4s)
		case dns.TypeAAAA:
			s := answer.(*dns.AAAA).AAAA.String()
			return IsInArray(s, j.ipv6s)
		default:
			return false
		}
	}

	if len(reply.Ns) > 0 {
		answer := reply.Ns[0]
		s := answer.String()
		return IsContains(s, j.domains)
	}

	return false
}

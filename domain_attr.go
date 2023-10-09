package chinadns

import (
	"github.com/0990/chinadns/pkg/util"
	"github.com/miekg/dns"
)

type DomainAttr int

const (
	DomainAttr_Invalid DomainAttr = iota
	DomainAttr_NO_CNAME
	DomainAttr_NO_IPV4
	DomainAttr_NO_IPV6
	DomainAttr_NO_HTTPS
)

func (s *Server) isAboardResolver(r *Resolver) bool {
	for _, v := range s.DNSAbroadServers {
		if r == v {
			return true
		}
	}
	return false
}

func (s *Server) getResolverAttr(resolver *Resolver) (ret []DomainAttr) {
	if resolver == nil {
		return nil
	}
	if s.isAboardResolver(resolver) {
		return s.DNSAbroadAttr
	}

	return nil
}

func filterLookupRetByAttrs(ret *LookupResult, attrs []DomainAttr) bool {
	if ret == nil {
		return false
	}

	if ret.reply.Question[0].Qtype == dns.TypeAAAA && util.IsInArray(DomainAttr_NO_IPV6, attrs) {
		ret.reply = genEmptyNoError(ret.reply)
		return true
	}

	var filter bool
	for _, v := range attrs {
		if filterLookupRetByAttr(ret, v) {
			filter = true
		}
	}

	return filter
}

func filterLookupRetByAttr(ret *LookupResult, attr DomainAttr) bool {
	if ret == nil {
		return false
	}

	if attr == DomainAttr_Invalid {
		return false
	}

	var filter bool
	for i := 0; i < len(ret.reply.Answer); {
		if isDel(ret.reply.Answer[i], attr) {
			ret.reply.Answer = append(ret.reply.Answer[:i], ret.reply.Answer[i+1:]...)
			filter = true
		} else {
			i++
		}
	}

	for _, v := range ret.reply.Answer {
		if filterHTTPSNoipv6(v, attr) {
			filter = true
		}
	}
	return filter
}

func isDel(rr dns.RR, attr DomainAttr) bool {
	switch rr.(type) {
	case *dns.CNAME:
		return attr == DomainAttr_NO_CNAME
	case *dns.A:
		return attr == DomainAttr_NO_IPV4
	case *dns.AAAA:
		return attr == DomainAttr_NO_IPV6
	case *dns.HTTPS:
		return attr == DomainAttr_NO_HTTPS
	default:
		return false
	}
}

func filterHTTPSNoipv6(rr dns.RR, attr DomainAttr) bool {
	switch a := rr.(type) {
	case *dns.HTTPS:
		if attr == DomainAttr_NO_IPV6 {
			return removeIPv6Hints(a)
		}
		return false
	default:
		return false
	}
}

func toDomainAttr(s string) DomainAttr {
	switch s {
	case "nocname":
		return DomainAttr_NO_CNAME
	case "noipv4":
		return DomainAttr_NO_IPV4
	case "noipv6":
		return DomainAttr_NO_IPV6
	case "nohttps":
		return DomainAttr_NO_HTTPS
	default:
		return DomainAttr_Invalid
	}
}

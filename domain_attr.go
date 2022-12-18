package chinadns

import (
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
	var del bool
	for i := 0; i < len(ret.reply.Answer); {
		if domainAttr2RType(attr) == ret.reply.Answer[i].Header().Rrtype {
			ret.reply.Answer = append(ret.reply.Answer[:i], ret.reply.Answer[i+1:]...)
			del = true
		} else {
			i++
		}
	}
	return del
}

func domainAttr2RType(attr DomainAttr) uint16 {
	switch attr {
	case DomainAttr_NO_CNAME:
		return dns.TypeCNAME
	case DomainAttr_NO_IPV4:
		return dns.TypeA
	case DomainAttr_NO_IPV6:
		return dns.TypeAAAA
	case DomainAttr_NO_HTTPS:
		return dns.TypeHTTPS
	default:
		return dns.TypeNone
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

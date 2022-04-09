package chinadns

import (
	"github.com/miekg/dns"
	"strings"
)

type DomainAttr int

const (
	DomainAttr_Invalid DomainAttr = iota
	DomainAttr_NO_CNAME
	DomainAttr_NO_IPV4
	DomainAttr_NO_IPV6
)

func (s *Server) getDomainAttr(domain string) (ret []DomainAttr) {
	attr, ok := s.Domain2Attr.Load(domain)
	if !ok {
		return nil
	}

	attrs := strings.Split(attr.(string), ";")
	for _, v := range attrs {
		ret = append(ret, toDomainAttr(v))
	}
	return ret
}

func filterLookupRetByAttrs(ret *LookupResult, attrs []DomainAttr) {
	if ret == nil {
		return
	}

	for _, v := range attrs {
		filterLookupRetByAttr(ret, v)
	}
}

func filterLookupRetByAttr(ret *LookupResult, attr DomainAttr) {
	if ret == nil {
		return
	}

	if attr == DomainAttr_Invalid {
		return
	}

	for i := 0; i < len(ret.reply.Answer); {
		if domainAttr2RType(attr) == ret.reply.Answer[i].Header().Rrtype {
			ret.reply.Answer = append(ret.reply.Answer[:i], ret.reply.Answer[i+1:]...)
		} else {
			i++
		}
	}
}

func domainAttr2RType(attr DomainAttr) uint16 {
	switch attr {
	case DomainAttr_NO_CNAME:
		return dns.TypeCNAME
	case DomainAttr_NO_IPV4:
		return dns.TypeA
	case DomainAttr_NO_IPV6:
		return dns.TypeAAAA
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
	default:
		return DomainAttr_Invalid
	}
}

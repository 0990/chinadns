package chinadns

import (
	"github.com/miekg/dns"
	"slices"
)

const retryNoError = 60

func removeIPv6Hints(rr *dns.HTTPS) bool {
	length := len(rr.Value)
	rr.Value = slices.DeleteFunc(rr.Value, func(kv dns.SVCBKeyValue) (del bool) {
		_, ok := kv.(*dns.SVCBIPv6Hint)

		return ok
	})

	return len(rr.Value) < length
}

func genEmptyNoError(request *dns.Msg) *dns.Msg {
	return GenEmptyMessage(request, dns.RcodeSuccess, retryNoError)
}

func GenEmptyMessage(request *dns.Msg, rCode int, retry uint32) *dns.Msg {
	resp := dns.Msg{}
	resp.SetRcode(request, rCode)
	resp.RecursionAvailable = true
	resp.Ns = genSOA(request, retry)
	return &resp
}

func genSOA(request *dns.Msg, retry uint32) []dns.RR {
	zone := ""
	if len(request.Question) > 0 {
		zone = request.Question[0].Name
	}

	soa := dns.SOA{
		// values copied from verisign's nonexistent .com domain
		// their exact values are not important in our use case because they are used for domain transfers between primary/secondary DNS servers
		Refresh: 1800,
		Retry:   retry,
		Expire:  604800,
		Minttl:  86400,
		// copied from  DNS
		Ns:     "fake-for-negative-caching.chinadns.com.",
		Serial: 100500,
		// rest is request-specific
		Hdr: dns.RR_Header{
			Name:   zone,
			Rrtype: dns.TypeSOA,
			Ttl:    10,
			Class:  dns.ClassINET,
		},
	}
	soa.Mbox = "hostmaster."
	if len(zone) > 0 && zone[0] != '.' {
		soa.Mbox += zone
	}
	return []dns.RR{&soa}
}

package chinadns

import (
	"github.com/miekg/dns"
	"sync"
	"time"
)

type DNSCache interface {
	Set(domain string, msg *dns.Msg)
	Get(domain string) (*dns.Msg, bool)
	Len() int
}

const DNSCache_TriggerGCCount = 1000

type dnsCache struct {
	sync.RWMutex
	cache     map[string]dnsCacheV
	expireSec int64
}

type dnsCacheV struct {
	msg         dns.Msg
	createdTime time.Time
}

func newDNSCache(expireSec int64) DNSCache {
	if expireSec == 0 {
		return &dnsCacheNone{}
	}

	return &dnsCache{
		RWMutex:   sync.RWMutex{},
		cache:     make(map[string]dnsCacheV),
		expireSec: expireSec,
	}
}

func (p *dnsCache) Set(domain string, msg *dns.Msg) {
	p.set(domain, msg)
	p.checkGC()
}

func (p *dnsCache) set(domain string, msg *dns.Msg) {
	p.Lock()
	defer p.Unlock()

	p.cache[domain] = dnsCacheV{
		msg:         *msg,
		createdTime: time.Now(),
	}
}

func (p *dnsCache) Get(domain string) (*dns.Msg, bool) {
	p.RLock()
	defer p.RUnlock()

	v, ok := p.cache[domain]
	if !ok {
		return nil, false
	}

	if p.isExpire(v, time.Now()) {
		return nil, false
	}

	return &v.msg, true
}

func (p *dnsCache) Len() int {
	p.Lock()
	defer p.Unlock()
	return len(p.cache)
}

func (p *dnsCache) checkGC() {
	p.Lock()
	defer p.Unlock()

	if len(p.cache) < DNSCache_TriggerGCCount {
		return
	}

	now := time.Now()
	for k, v := range p.cache {
		if p.isExpire(v, now) {
			delete(p.cache, k)
		}
	}
}

func (p *dnsCache) isExpire(dnsCacheV dnsCacheV, now time.Time) bool {
	if now.Sub(dnsCacheV.createdTime)/time.Second > time.Duration(p.expireSec) {
		return true
	}
	return false
}

type dnsCacheNone struct {
}

func (p *dnsCacheNone) Set(domain string, msg *dns.Msg) {
	return
}

func (p *dnsCacheNone) Get(domain string) (*dns.Msg, bool) {
	return nil, false
}

func (p *dnsCacheNone) Len() int {
	return 0
}

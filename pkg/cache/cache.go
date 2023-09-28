package cache

import (
	"github.com/miekg/dns"
	"sync"
	"time"
)

type DNSCache interface {
	Set(q dns.Question, msg any)
	Get(q dns.Question) (any, bool)
	Len() int
}

const DNSCache_TriggerGCCount = 1000

type dnsCache struct {
	sync.RWMutex
	cache     map[dns.Question]dnsCacheV
	expireSec int64
}

type dnsCacheV struct {
	lr          any
	createdTime time.Time
}

func NewDNSCache(expireSec int64) DNSCache {
	if expireSec <= 0 {
		return &dnsCacheNone{}
	}

	return &dnsCache{
		RWMutex:   sync.RWMutex{},
		cache:     make(map[dns.Question]dnsCacheV),
		expireSec: expireSec,
	}
}

func (p *dnsCache) Set(q dns.Question, lr any) {
	p.set(q, lr)
	p.checkGC()
}

func (p *dnsCache) set(q dns.Question, lr any) {
	p.Lock()
	defer p.Unlock()

	p.cache[q] = dnsCacheV{
		lr:          lr,
		createdTime: time.Now(),
	}
}

func (p *dnsCache) Get(q dns.Question) (any, bool) {
	p.RLock()
	defer p.RUnlock()

	v, ok := p.cache[q]
	if !ok {
		return nil, false
	}

	if p.isExpire(v, time.Now()) {
		return nil, false
	}

	return v.lr, true
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

func (p *dnsCacheNone) Set(q dns.Question, lr any) {
	return
}

func (p *dnsCacheNone) Get(q dns.Question) (any, bool) {
	return nil, false
}

func (p *dnsCacheNone) Len() int {
	return 0
}

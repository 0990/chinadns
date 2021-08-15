package chinadns

import (
	"fmt"
	"github.com/miekg/dns"
	"testing"
	"time"
)

func TestDNSCache_Get(t *testing.T) {
	c := newDNSCache(1)
	domain := "hello"
	msg := &dns.Msg{}
	c.Set(domain, msg)
	msg1, ok := c.Get(domain)
	if !ok {
		t.Fatal("get fail")
	}

	if msg1 != msg {
		t.Fatal("cache v not equal")
	}

	time.Sleep(time.Second * 3)

	_, ok = c.Get(domain)
	if ok {
		t.Fatal("cache still exist,should expire")
	}

}

func TestDNSCache_GC(t *testing.T) {
	batch1Count := 10
	batch2Count := DNSCache_TriggerGCCount
	c := newDNSCache(1)

	domainIndex := 0
	newDomain := func() string {
		domainIndex++
		return fmt.Sprintf("%v", domainIndex)
	}
	for i := 0; i < batch1Count; i++ {
		c.Set(newDomain(), &dns.Msg{})
	}

	time.Sleep(time.Second * 3)

	for i := 0; i < batch2Count; i++ {
		c.Set(newDomain(), &dns.Msg{})
	}

	if c.Len() != batch2Count {
		t.Fail()
	}
}

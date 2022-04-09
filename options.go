package chinadns

import (
	"bufio"
	"fmt"
	"github.com/0990/chinadns/matcher"
	"github.com/yl2chen/cidranger"
	"net"
	"os"
	"sync"
)

// ServerOption provides ChinaDNS server options. Please use WithXXX functions to generate Options.
type ServerOption func(*serverOptions) error

type serverOptions struct {
	Listen string // Listening address, such as `[::]:53`, `0.0.0.0:53`

	CacheExpireSec int64

	Domain2IP   sync.Map
	Domain2Attr sync.Map

	DNSChinaServers  resolverList // DNS servers which can be trusted
	DNSAbroadServers resolverList // DNS servers which may return polluted results

	ChinaCIDR cidranger.Ranger

	chnDomainMatcher matcher.Matcher
	gfwDomainMatcher matcher.Matcher
}

func newServerOptions() *serverOptions {
	return &serverOptions{
		Listen: "[::]:53",
	}
}

func WithListenAddr(addr string) ServerOption {
	return func(o *serverOptions) error {
		o.Listen = addr
		return nil
	}
}

func WithCacheExpireSec(sec int) ServerOption {
	return func(o *serverOptions) error {
		o.CacheExpireSec = int64(sec)
		return nil
	}
}

func WithDomain2IP(domain2ip map[string]string) ServerOption {
	return func(o *serverOptions) error {
		for k, v := range domain2ip {
			o.Domain2IP.Store(k, v)
		}
		return nil
	}
}

func WithDomain2Attr(domain2attr map[string]string) ServerOption {
	return func(o *serverOptions) error {
		for k, v := range domain2attr {
			o.Domain2Attr.Store(k, v)
		}
		return nil
	}
}

func WithDNS(dnsChina, dnsAbroad []string) ServerOption {
	return func(o *serverOptions) error {
		for _, schema := range dnsChina {
			newResolver, err := ParseResolver(schema, false)
			if err != nil {
				return err
			}
			o.DNSChinaServers = uniqueAppendResolver(o.DNSChinaServers, newResolver)
		}

		for _, schema := range dnsAbroad {
			newResolver, err := ParseResolver(schema, false)
			if err != nil {
				return err
			}
			o.DNSAbroadServers = uniqueAppendResolver(o.DNSAbroadServers, newResolver)
		}

		return nil
	}
}

//func WithGFWFile(addr []string) ServerOption {
//	return func(o *serverOptions) error {
//		gfw, err := gfwlist.NewFromFiles(addr, false)
//		if err != nil {
//			return err
//		}
//		o.gfwlist = gfw
//		return nil
//	}
//}

func WithCHNFile(paths []string) ServerOption {
	return func(o *serverOptions) error {
		if len(paths) == 0 {
			return fmt.Errorf("empty for China route list")
		}

		for _, path := range paths {
			err := addCHNFile(o, path)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func addCHNFile(o *serverOptions, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("fail to open China route list: %w", err)

	}
	defer file.Close()

	if o.ChinaCIDR == nil {
		o.ChinaCIDR = cidranger.NewPCTrieRanger()
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		_, network, err := net.ParseCIDR(scanner.Text())
		if err != nil {
			return fmt.Errorf("parse %s as CIDR failed: %v", scanner.Text(), err.Error())
		}
		err = o.ChinaCIDR.Insert(cidranger.NewBasicRangerEntry(*network))
		if err != nil {
			return fmt.Errorf("insert %s as CIDR failed: %v", scanner.Text(), err.Error())
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("fail to scan china route list: %v", err.Error())
	}
	return nil
}

func WithChnDomain(path string) ServerOption {
	return func(o *serverOptions) error {
		if path == "" {
			return fmt.Errorf("empty for China domain list")
		}

		m, err := matcher.New("combine", path)
		if err != nil {
			return err
		}
		o.chnDomainMatcher = m
		return nil
	}
}

func WithGfwDomain(path string) ServerOption {
	return func(o *serverOptions) error {
		if path == "" {
			return fmt.Errorf("empty for Gfw domain list")
		}

		m, err := matcher.New("combine", path)
		if err != nil {
			return err
		}
		o.gfwDomainMatcher = m
		return nil
	}
}

func uniqueAppendString(to []string, item string) []string {
	for _, e := range to {
		if item == e {
			return to
		}
	}
	return append(to, item)
}

func uniqueAppendResolver(to []*Resolver, item *Resolver) []*Resolver {
	for _, e := range to {
		if item.GetAddr() == e.GetAddr() {
			return to
		}
	}
	return append(to, item)
}

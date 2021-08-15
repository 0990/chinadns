package chinadns

import (
	"bufio"
	"fmt"
	"github.com/0990/chinadns/gfwlist"
	"github.com/yl2chen/cidranger"
	"net"
	"os"
)

// ServerOption provides ChinaDNS server options. Please use WithXXX functions to generate Options.
type ServerOption func(*serverOptions) error

type serverOptions struct {
	Listen string // Listening address, such as `[::]:53`, `0.0.0.0:53`

	CacheExpireSec int64

	DNSChinaServers  resolverList // DNS servers which can be trusted
	DNSAbroadServers resolverList // DNS servers which may return polluted results

	gfwlist   *gfwlist.GFWList
	ChinaCIDR cidranger.Ranger
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

func WithGFWFile(addr []string) ServerOption {
	return func(o *serverOptions) error {
		gfw, err := gfwlist.NewFromFiles(addr, false)
		if err != nil {
			return err
		}
		o.gfwlist = gfw
		return nil
	}
}

func WithCHNFile(path string) ServerOption {
	return func(o *serverOptions) error {
		if path == "" {
			return fmt.Errorf("empty for China route list")
		}
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

package chinadns

import (
	"github.com/0990/chinadns/gfwlist"
)

// ServerOption provides ChinaDNS server options. Please use WithXXX functions to generate Options.
type ServerOption func(*serverOptions) error

type serverOptions struct {
	Listen string // Listening address, such as `[::]:53`, `0.0.0.0:53`

	DNSChinaServers  resolverList // DNS servers which can be trusted
	DNSAbroadServers resolverList // DNS servers which may return polluted results

	gfwlist *gfwlist.GFWList
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

func WithGFWFile(addr string) ServerOption {
	return func(o *serverOptions) error {
		gfw, err := gfwlist.NewFromFile("gfwlist.txt", false)
		if err != nil {
			return err
		}
		o.gfwlist = gfw
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

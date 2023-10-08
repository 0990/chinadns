package chinadns

import (
	"context"
	"errors"
	"fmt"
	"github.com/0990/chinadns/pkg/doh"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

type Client struct {
	*clientOptions

	UDPCli *dns.Client
	TCPCli *dns.Client
	DoHCli *doh.Client

	DoHCliProxy *doh.Client
	proxyProto  string
	proxyAddr   string
}

func NewClient(opts ...ClientOption) (*Client, error) {
	o := new(clientOptions)
	for _, f := range opts {
		f(o)
	}

	var proxyProto, proxyAddr string
	if o.DNSAbroadProxy != "" {
		attrs := strings.Split(o.DNSAbroadProxy, "://")
		if len(attrs) != 2 {
			return nil, fmt.Errorf("invalid proxy format")
		}
		proxyProto = attrs[0]
		proxyAddr = attrs[1]
		switch proxyProto {
		case "socks5":
		default:
			return nil, errors.New("not support proxy protocol")
		}
	}

	return &Client{
		clientOptions: o,
		UDPCli: &dns.Client{
			Net:     "udp",
			Timeout: o.Timeout,
		},
		TCPCli: &dns.Client{
			Net:     "tcp",
			Timeout: o.Timeout,
		},
		DoHCli: doh.NewClient(
			doh.WithTimeout(o.Timeout),
			doh.WithSkipQueryMySelf(true),
		),
		DoHCliProxy: doh.NewClient(
			doh.WithTimeout(o.Timeout),
			doh.WithSkipQueryMySelf(true),
			doh.WithSocks5Proxy(proxyAddr),
		),
		proxyProto: proxyProto,
		proxyAddr:  proxyAddr,
	}, nil
}

// lookupProxyPriority will try to use proxy first, if proxy failed, use normal method.
func (c *Client) lookupProxyPriority(ctx context.Context, req *dns.Msg, server *Resolver) (reply *dns.Msg, remark string, err error) {
	if len(c.proxyProto) == 0 {
		return c.lookup(ctx, req, server)
	}

	return c.lookupByProxy(ctx, req, server)
}

func (c *Client) lookup(ctx context.Context, req *dns.Msg, server *Resolver) (reply *dns.Msg, remark string, err error) {
	logger := logrus.WithFields(logrus.Fields{
		"question": questionString(&req.Question[0]),
		"dns":      server,
		"id":       reqID(req),
	})

	for _, protocol := range server.Protocols {
		switch protocol {
		case "udp":
			reply, _, err = c.UDPCli.ExchangeContext(ctx, req, server.GetAddr())
			if err == nil {
				return
			}

			if reply != nil && reply.Truncated {
				logger.Error("Truncated msg received.Conder enlarge your UDP max size")
			}
		case "tcp":
			reply, _, err = c.TCPCli.ExchangeContext(ctx, req, server.GetAddr())
			if err == nil {
				return
			}
			logger.WithError(err).Error("Fail to send TCP query.")
		case "doh":
			reply, _, err = c.DoHCli.Exchange(ctx, req, server.GetAddr())
			if err == nil {
				return
			}
			logger.WithError(err).Error("Fail to send DoH query.")
		default:
			logger.Errorf("Protocol %s is unsupported in normal method.", protocol)
			return
		}
	}
	return
}

func (c *Client) lookupByProxy(ctx context.Context, req *dns.Msg, server *Resolver) (reply *dns.Msg, remark string, err error) {
	logger := logrus.WithFields(logrus.Fields{
		"question": questionString(&req.Question[0]),
		"dns":      server,
		"id":       reqID(req),
	})

	remark = fmt.Sprintf("useproxy:%s://%s", c.proxyProto, c.proxyAddr)

	for _, protocol := range server.Protocols {
		switch protocol {
		case "udp":
			reply, err = lookUpBySocks5(ctx, c.proxyAddr, protocol, server.GetAddr(), req)
			if err == nil {
				return
			}

			if reply != nil && reply.Truncated {
				logger.Error("Truncated msg received.Conder enlarge your UDP max size")
			}
		case "tcp":
			reply, err = lookUpBySocks5(ctx, c.proxyAddr, protocol, server.GetAddr(), req)
			if err == nil {
				return
			}
			logger.WithError(err).Error("Fail to send TCP query.")
		case "doh":
			reply, _, err = c.DoHCliProxy.Exchange(ctx, req, server.GetAddr())
			if err == nil {
				return
			}
			logger.WithError(err).Error("Fail to send DoH query.")
		default:
			logger.Errorf("Protocol %s is unsupported in normal method.", protocol)
			return
		}
	}
	return
}

type clientOptions struct {
	Timeout        time.Duration // Timeout for one DNS query
	UDPMaxSize     int           // Max message size for UDP queries
	TCPOnly        bool          // Use TCP only
	DNSAbroadProxy string        //socks5://x.x.x.x:port
}

type ClientOption func(*clientOptions)

func WithTimeout(t time.Duration) ClientOption {
	return func(o *clientOptions) {
		o.Timeout = t
	}
}

func WithUDPMaxBytes(max int) ClientOption {
	return func(o *clientOptions) {
		o.UDPMaxSize = max
	}
}

func WithTCPOnly(b bool) ClientOption {
	return func(o *clientOptions) {
		o.TCPOnly = b
	}
}

func WithDNSAboardProxy(proxy string) ClientOption {
	return func(o *clientOptions) {
		o.DNSAbroadProxy = proxy
	}
}

package chinadns

import (
	"github.com/0990/chinadns/doh"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"time"
)

type Client struct {
	*clientOptions
	UDPCli *dns.Client
	TCPCli *dns.Client
	DoHCli *doh.Client
}

func NewClient(opts ...ClientOption) *Client {
	o := new(clientOptions)
	for _, f := range opts {
		f(o)
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
	}
}

func (c *Client) lookup(reqID uint32, req *dns.Msg, server *Resolver) (reply *dns.Msg, rtt time.Duration, err error) {
	logger := logrus.WithFields(logrus.Fields{
		"question": questionString(&req.Question[0]),
		"server":   server,
		"id":       reqID,
	})

	var rtt0 time.Duration

	for _, protocol := range server.Protocols {
		switch protocol {
		case "udp":
			//logger.Debug("Query upstream udp")
			reply, rtt0, err = c.UDPCli.Exchange(req, server.GetAddr())
			rtt += rtt0
			if err == nil {
				return
			}

			if reply != nil && reply.Truncated {
				logger.Error("Truncated msg received.Conder enlarge your UDP max size")
			}
		case "tcp":
			logger.Debug("Query upstream tcp")
			reply, rtt0, err = c.TCPCli.Exchange(req, server.GetAddr())
			if err == nil {
				return
			}
			rtt += rtt0
			logger.WithError(err).Error("Fail to send TCP query.")
		case "doh":
			logger.Debug("Query upstream doh")
			reply, rtt, err = c.DoHCli.Exchange(req, server.GetAddr())
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
	Timeout    time.Duration // Timeout for one DNS query
	UDPMaxSize int           // Max message size for UDP queries
	TCPOnly    bool          // Use TCP only
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

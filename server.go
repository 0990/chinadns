package chinadns

import (
	"context"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	*serverOptions
	*Client
	UDPServer *dns.Server
	TCPServer *dns.Server

	requestID uint32

	cache DNSCache
}

func NewServer(cli *Client, opts ...ServerOption) (*Server, error) {
	var o = newServerOptions()

	for _, f := range opts {
		if err := f(o); err != nil {
			return nil, err
		}
	}

	//ss-tproxy need 53 port udp4,otherwise not work,set udp back when fix
	s := &Server{
		serverOptions: o,
		Client:        cli,
		UDPServer:     &dns.Server{Addr: o.Listen, Net: "udp", ReusePort: true},
		TCPServer:     &dns.Server{Addr: o.Listen, Net: "tcp", ReusePort: true},
		cache:         newDNSCache(o.CacheExpireSec),
	}

	s.UDPServer.Handler = dns.HandlerFunc(s.Serve)
	s.TCPServer.Handler = dns.HandlerFunc(s.Serve)

	return s, nil
}

func (s *Server) Run() error {
	logrus.Info("Start server at ", s.Listen)
	eg, _ := errgroup.WithContext(context.Background())
	eg.Go(func() error {
		return runUDPServer(s.UDPServer)
	})
	eg.Go(func() error {
		return runTCPServer(s.TCPServer)
	})
	return eg.Wait()
}

func (s *Server) normalizeRequest(req *dns.Msg) {
	req.RecursionDesired = true
	if !s.TCPOnly {
		setUDPSize(req, uint16(s.UDPMaxSize))
	}
}

func setUDPSize(req *dns.Msg, size uint16) uint16 {
	if size <= dns.MinMsgSize {
		return dns.MinMsgSize
	}
	// https://tools.ietf.org/html/rfc6891#section-6.2.5
	if e := req.IsEdns0(); e != nil {
		if e.UDPSize() >= size {
			return e.UDPSize()
		}
		e.SetUDPSize(size)
		return size
	}
	req.SetEdns0(size, false)
	return size
}

package chinadns

import "github.com/miekg/dns"

func runUDPServer(s *dns.Server) error {
	pc, err := listenUDP(s.Net, s.Addr, s.ReusePort)
	if err != nil {
		return err
	}

	s.PacketConn = pc
	return s.ActivateAndServe()
}

func runTCPServer(s *dns.Server) error {
	l, err := listenTCP(s.Net, s.Addr, s.ReusePort)
	if err != nil {
		return err
	}

	s.Listener = l
	return s.ActivateAndServe()
}

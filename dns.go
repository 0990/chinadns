package chinadns

import (
	"errors"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"time"
)

type LookupFunc func(request *dns.Msg, server *Resolver) (reply *dns.Msg, rtt time.Duration, err error)

func (s *Server) Serve(w dns.ResponseWriter, req *dns.Msg) {
	logger := logrus.WithField("question", questionString(&req.Question[0]))

	start := time.Now()

	s.normalizeRequest(req)

	qName := req.Question[0].Name

	dnsServer := s.DNSChinaServers
	if s.gfwlist.IsDomainBlocked(qName) {
		dnsServer = s.DNSAbroadServers
	}

	logger.Debugf("%s use dns server:%s", qName, dnsServer)

	reply, err := lookupInServers(req, dnsServer, time.Second*3, s.lookup)
	if err != nil {
		logger.WithError(err).Error("query error")

		reply = new(dns.Msg)
		reply.SetReply(req)
		_ = w.WriteMsg(reply)
		return
	}

	// https://github.com/miekg/dns/issues/216
	reply.Compress = true
	_ = w.WriteMsg(reply)

	logger.Debug("SERVING RTT: ", time.Since(start))
}

func lookupInServers(req *dns.Msg, servers []*Resolver, waitInterval time.Duration, lookup LookupFunc) (*dns.Msg, error) {
	if len(servers) == 0 {
		return nil, errors.New("no servers")
	}

	logger := logrus.WithField("question", questionString(&req.Question[0]))

	queryNext := make(chan struct{}, len(servers))
	queryNext <- struct{}{}

	result := make(chan *dns.Msg, 1)

	dolookup := func(server *Resolver) {
		logger := logger.WithField("server", server.GetAddr())

		reply, rtt, err := lookup(req.Copy(), server)
		if err != nil {
			queryNext <- struct{}{}
			return
		}

		select {
		case result <- reply:
			logger.Debug("Query RTT: ", rtt)
		default:

		}
	}

	waitTicker := time.NewTimer(waitInterval)
	defer waitTicker.Stop()

	for _, server := range servers {
		waitTicker.Reset(waitInterval)
		select {
		case <-queryNext:
			go dolookup(server)
		case <-waitTicker.C:
			go dolookup(server)
		}
	}

	select {
	case ret := <-result:
		return ret, nil
	case <-time.After(time.Second * 10):
		return nil, errors.New("query timeout")
	}
}

func questionString(q *dns.Question) string {
	return q.Name + " " + dns.TypeToString[q.Qtype]
}

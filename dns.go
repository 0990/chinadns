package chinadns

import (
	"errors"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"net"
	"sync/atomic"
	"time"
)

type LookupFunc func(id uint32, request *dns.Msg, server *Resolver) (reply *dns.Msg, rtt time.Duration, err error)

func (s *Server) Serve(w dns.ResponseWriter, req *dns.Msg) {
	reqID := s.newID()

	logger := logrus.WithFields(logrus.Fields{
		"question": questionString(&req.Question[0]),
		"id":       reqID,
	})

	start := time.Now()

	var reply *dns.Msg
	var err error

	defer func() {
		if reply == nil {
			reply = new(dns.Msg)
			reply.SetReply(req)
		}
		// https://github.com/miekg/dns/issues/216
		reply.Compress = true
		_ = w.WriteMsg(reply)
		logger.Debug("SERVING RTT: ", time.Since(start), " IP:", answerIPString(reply))
	}()

	reqDomain := reqDomain(req)

	if v, ok := s.cache.Get(reqDomain); ok {
		reply = v
		reply.Id = req.Id
		logger.Debug("Cache HIT")
		return
	}

	s.normalizeRequest(req)

	//gfw block的域名直接使用国外dns
	if s.gfwlist.IsDomainBlocked(reqDomain) {
		reply, err = lookupInServers(reqID, req, s.DNSAbroadServers, time.Second*3, s.lookup)
		if err != nil {
			logger.WithError(err).Error("query error")
			return
		}
		return
	}

	replyAbroad := make(chan *dns.Msg, 1)
	go func() {
		reply, err := lookupInServers(reqID, req, s.DNSAbroadServers, time.Second*3, s.lookup)
		if err != nil {
			logger.WithError(err).Error("query error")
			return
		}
		replyAbroad <- reply
	}()

	reply, err = lookupInServers(reqID, req, s.DNSChinaServers, time.Second*3, s.lookup)
	if err != nil {
		logger.WithError(err).Error("query error")
		return
	}

	//使用国内dns但返回的是国外ip,则用国外dns的查询结果
	if !s.isReplyIPChn(reply) {
		logger.Debug("ChinaDNS SERVING RTT: ", time.Since(start), " IP:", answerIPString(reply))

		logrus.WithField("domain", reqDomain).Warn("use china dns,but reply is abroad")
		select {
		case reply = <-replyAbroad:
			return
		case <-time.After(time.Second * 3):
			return
		}
	}

	if reply != nil {
		s.cache.Set(reqDomain, reply)
	}
}

func reqDomain(request *dns.Msg) string {
	qName := request.Question[0].Name

	domain := qName
	if len(domain) > 0 {
		domain = domain[:len(domain)-1]
	}
	return domain
}

func (s *Server) newID() uint32 {
	return atomic.AddUint32(&s.requestID, 1)
}

func (s *Server) isReplyIPChn(reply *dns.Msg) bool {
	for _, ip := range replyIP(reply) {
		ok, err := s.ChinaCIDR.Contains(ip)
		if err != nil {
			logrus.WithError(err).WithField("ip", ip.String()).Error("ChinaCIDR.Contains")
			return false
		}
		if ok {
			return true
		}
	}
	return false
}

func replyIP(reply *dns.Msg) []net.IP {
	var ip []net.IP
	for _, rr := range reply.Answer {
		switch answer := rr.(type) {
		case *dns.A:
			ip = append(ip, answer.A)
		case *dns.AAAA:
			ip = append(ip, answer.AAAA)
		case *dns.CNAME:
			continue
		default:
			continue
		}
	}
	return ip
}

func answerIPString(reply *dns.Msg) string {
	ips := replyIP(reply)
	var ip string
	for _, v := range ips {
		ip += v.String()
		ip += ";"
	}
	return ip
}

func lookupInServers(reqID uint32, req *dns.Msg, servers []*Resolver, waitInterval time.Duration, lookup LookupFunc) (*dns.Msg, error) {
	if len(servers) == 0 {
		return nil, errors.New("no servers")
	}

	//logger := logrus.WithField("question", questionString(&req.Question[0]))

	queryNext := make(chan struct{}, len(servers))
	queryNext <- struct{}{}

	result := make(chan *dns.Msg, 1)

	dolookup := func(server *Resolver) {
		//logger := logger.WithField("server", server.GetAddr())

		reply, _, err := lookup(reqID, req.Copy(), server)
		if err != nil {
			queryNext <- struct{}{}
			return
		}

		select {
		case result <- reply:
			//logger.Debug("Query RTT: ", rtt)
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
	case <-time.After(time.Second * 5):
		return nil, errors.New("query timeout")
	}
}

func questionString(q *dns.Question) string {
	return q.Name + " " + dns.TypeToString[q.Qtype]
}

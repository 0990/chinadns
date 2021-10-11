package chinadns

import (
	"errors"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"net"
	"strings"
	"sync/atomic"
	"time"
)

type LookupFunc func(id uint32, request *dns.Msg, server *Resolver) (reply *dns.Msg, rtt time.Duration, err error)

type LookupResult struct {
	reply    *dns.Msg
	resolver *Resolver
}

func (s *Server) Serve(w dns.ResponseWriter, req *dns.Msg) {
	reqID := s.newID()

	logger := logrus.WithFields(logrus.Fields{
		"question": questionString(&req.Question[0]),
		"id":       reqID,
	})

	start := time.Now()

	var err error
	var lookupRet *LookupResult

	//var fromCache bool

	reqDomain := reqDomain(req)

	defer func() {
		//if !fromCache && lookupRet != nil {
		//	s.cache.Set(reqDomain, lookupRet.reply)
		//}
		if lookupRet == nil {
			logger.Warn("reply==nil")
			reply := new(dns.Msg)
			reply.SetReply(req)

			lookupRet = &LookupResult{reply: reply}
		}
		// https://github.com/miekg/dns/issues/216
		lookupRet.reply.Compress = true
		_ = w.WriteMsg(lookupRet.reply)

		replyRet := replyString(lookupRet.reply)

		logger.WithFields(logrus.Fields{
			"RTT":    timeSinceMS(start),
			"server": lookupRet.resolver,
			"reply":  replyRet,
			"detail": lookupRet.reply.String(),
		}).Debug("DNS reply")

		if replyRet == "" {
			logger.WithFields(logrus.Fields{
				"RTT":    timeSinceMS(start),
				"server": lookupRet.resolver,
				"reply":  replyRet,
			}).Error("DNS reply empty")
		}
	}()

	//if v, ok := s.cache.Get(reqDomain); ok {
	//	fromCache = true
	//	reply = v
	//	reply.Id = req.Id
	//	logger.Debug("Cache HIT")
	//	return
	//}

	s.normalizeRequest(req)

	//gfw block的域名直接使用国外dns
	if s.gfwlist.IsDomainBlocked(reqDomain) {
		lookupRet, err = lookupInServers(reqID, req, s.DNSAbroadServers, time.Second*2, s.lookup)
		if err != nil {
			logger.WithError(err).Error("query error")
			return
		}

		logger.WithFields(logrus.Fields{
			"RTT":    timeSinceMS(start),
			"server": lookupRet.resolver,
			"reply":  replyString(lookupRet.reply),
		}).Debug("Query result")
		return
	}

	lookupRetAbroad := make(chan *LookupResult, 1)
	go func() {
		ret, err := lookupInServers(reqID, req, s.DNSAbroadServers, time.Second*2, s.lookup)
		if err != nil {
			logger.WithError(err).Error("query error")
			return
		}
		lookupRetAbroad <- ret
	}()

	lookupRet, err = lookupInServers(reqID, req, s.DNSChinaServers, time.Millisecond*500, s.lookup)
	if err != nil {
		logger.WithError(err).Error("query error")
		return
	}

	logger.WithFields(logrus.Fields{
		"RTT":    timeSinceMS(start),
		"server": lookupRet.resolver,
		"reply":  replyString(lookupRet.reply),
	}).Debug("Query result")

	//使用国内dns但返回的是国外ip,则用国外dns的查询结果
	if !s.isReplyIPChn(lookupRet.reply) {
		logrus.WithField("domain", reqDomain).Warn("use china dns,but reply is abroad")
		select {
		case lookupRet = <-lookupRetAbroad:
			logger.WithFields(logrus.Fields{
				"RTT":    timeSinceMS(start),
				"server": lookupRet.resolver,
				"reply":  replyString(lookupRet.reply),
			}).Debug("Query result")
			return
		case <-time.After(time.Second * 3):
			return
		}
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

func replyCDName(reply *dns.Msg) (ret []string) {
	for _, rr := range reply.Answer {
		switch answer := rr.(type) {
		case *dns.A:
			continue
		case *dns.AAAA:
			continue
		case *dns.CNAME:
			ret = append(ret, answer.Target)
		case *dns.DNAME:
			ret = append(ret, answer.Target)
		default:
			continue
		}
	}
	return ret
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

func answerCDNameString(reply *dns.Msg) string {
	ips := replyCDName(reply)
	return strings.Join(ips, ";")
}

func replyString(reply *dns.Msg) string {
	//return reply.String()
	return answerIPString(reply) + answerCDNameString(reply)
}

func lookupInServers(reqID uint32, req *dns.Msg, servers []*Resolver, waitInterval time.Duration, lookup LookupFunc) (*LookupResult, error) {
	if len(servers) == 0 {
		return nil, errors.New("no servers")
	}

	result := make(chan *LookupResult, 1)

	doLookup := func(server *Resolver) {
		reply, _, err := lookup(reqID, req.Copy(), server)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"server":   server,
				"question": questionString(&req.Question[0]),
			}).WithError(err).Error("lookup")
			return
		}

		if replyString(reply) == "" {
			return
		}

		select {
		case result <- &LookupResult{
			reply:    reply,
			resolver: server,
		}:
		default:
		}
	}

	for _, server := range servers {
		go doLookup(server)
	}

	select {
	case ret := <-result:
		return ret, nil
	case <-time.After(waitInterval):
		return nil, errors.New("query timeout")
	}
}

func questionString(q *dns.Question) string {
	return q.Name + " " + dns.TypeToString[q.Qtype]
}

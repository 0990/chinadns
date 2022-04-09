package chinadns

import (
	"errors"
	"fmt"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"net"
	"strings"
	"sync"
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
		"q":  questionString(&req.Question[0]),
		"id": reqID,
	})

	start := time.Now()

	var err error
	var lookupRet *LookupResult

	var hitCache bool

	question := req.Question[0]

	reqDomain := reqDomain(req)

	defer func() {
		if attrs := s.getDomainAttr(reqDomain); len(attrs) > 0 {
			filterLookupRetByAttrs(lookupRet, attrs)
		}

		if !hitCache && lookupRet != nil {
			replyRet := replyString(lookupRet.reply)
			if replyRet != "" {
				s.cache.Set(question, lookupRet)
			}
		}

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
			"RTT":      timeSinceMS(start),
			"resolver": lookupRet.resolver,
			"reply":    replyRet,
			"z":        lookupRet.reply.String(),
			"hitCache": hitCache,
		}).Debug("DNS reply")
	}()

	//自定义域名中查找
	if reply, ok := s.lookUpInCustom(reqID, reqDomain, req); ok {
		lookupRet = &LookupResult{
			reply:    reply,
			resolver: nil,
		}
		return
	}

	if v, ok := s.cache.Get(question); ok {
		hitCache = true

		reply := v.reply.Copy()
		reply.Id = req.Id
		lookupRet = &LookupResult{
			reply:    reply,
			resolver: v.resolver,
		}
		return
	}

	//s.normalizeRequest(req)

	//国内域名直接走国内dns
	if s.chnDomainMatcher.IsMatch(reqDomain) {
		lookupRet, err = lookupInServers(reqID, req, s.DNSChinaServers, time.Second*2, s.lookup)
		if err != nil {
			logger.WithError(err).Error("query error")
			return
		}

		logger.WithFields(logrus.Fields{
			"RTT":      timeSinceMS(start),
			"resolver": lookupRet.resolver,
			"reply":    replyString(lookupRet.reply),
		}).Debug("Query result")
		return
	}

	//gfw block的域名直接使用国外dns
	if s.gfwDomainMatcher.IsMatch(reqDomain) {
		lookupRet, err = lookupInServers(reqID, req, s.DNSAbroadServers, time.Second*2, s.lookup)
		if err != nil {
			logger.WithError(err).Error("query error")
			return
		}

		logger.WithFields(logrus.Fields{
			"RTT":      timeSinceMS(start),
			"resolver": lookupRet.resolver,
			"reply":    replyString(lookupRet.reply),
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

	lookupRet, err = lookupInServers(reqID, req, s.DNSChinaServers, time.Second*1, s.lookup)
	if err != nil {
		logger.WithError(err).Error("query error")
		return
	}

	logger.WithFields(logrus.Fields{
		"RTT":      timeSinceMS(start),
		"resolver": lookupRet.resolver,
		"reply":    replyString(lookupRet.reply),
	}).Debug("Query result")

	//使用国内dns但返回的是国外ip,则用国外dns的查询结果
	if !s.isReplyIPChn(lookupRet.reply) {
		logrus.WithField("domain", reqDomain).Warn("use china dns,but reply is abroad")
		select {
		case lookupRet = <-lookupRetAbroad:
			logger.WithFields(logrus.Fields{
				"RTT":      timeSinceMS(start),
				"resolver": lookupRet.resolver,
				"reply":    replyString(lookupRet.reply),
			}).Debug("Query result")
			return
		case <-time.After(time.Second * 3):
			return
		}
	}
}

func isIPV4(ip string) bool {
	return net.ParseIP(ip).To4() != nil
}

func isIPV6(ip string) bool {
	return net.ParseIP(ip).To16() != nil
}

func getIPV4(vs []string) (ret []string) {
	for _, v := range vs {
		if isIPV4(v) {
			ret = append(ret, v)
		}
	}
	return
}

func getIPV6(vs []string) (ret []string) {
	for _, v := range vs {
		if isIPV6(v) {
			ret = append(ret, v)
		}
	}
	return
}

//查找自定义域名
func (s *Server) lookUpInCustom(reqID uint32, domain string, req *dns.Msg) (*dns.Msg, bool) {
	ret, ok := s.Domain2IP.Load(domain)
	if !ok {
		return nil, false
	}

	logger := logrus.WithFields(logrus.Fields{
		"q":  questionString(&req.Question[0]),
		"id": reqID,
	})

	qType := req.Question[0].Qtype

	allIPs := strings.Split(ret.(string), ";")

	var useIPs []string
	var format string
	switch qType {
	case dns.TypeA:
		useIPs = getIPV4(allIPs)
		format = "%s. IN 3600 A %s"
	case dns.TypeAAAA:
		useIPs = getIPV6(allIPs)
		format = "%s. IN 3600 AAAA %s"
	default:
		return nil, false
	}

	if len(useIPs) == 0 {
		return nil, false
	}

	var rrs []dns.RR

	if !isAnswerNil(useIPs) {
		for _, ip := range useIPs {
			s := fmt.Sprintf(format, domain, ip)
			rr, err := dns.NewRR(s)
			if err != nil {
				logger.WithField("rr", s).WithError(err).Error("dns.NewRR")
				return nil, false
			}
			rrs = append(rrs, rr)
		}
	}

	reply := new(dns.Msg)
	reply.SetReply(req)
	reply.Answer = rrs

	return reply, true
}

func isAnswerNil(ips []string) bool {
	if len(ips) != 1 {
		return false
	}

	ip := ips[0]
	switch ip {
	case "::", "0:0:0:0:0:0:0:0", "0.0.0.0":
		return true
	default:
		return false
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
	return true
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

	var wg sync.WaitGroup

	doLookup := func(server *Resolver) {
		defer wg.Done()
		reply, rtt, err := lookup(reqID, req.Copy(), server)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"resolver": server,
				"question": questionString(&req.Question[0]),
				"RTT":      int64(rtt / time.Millisecond),
			}).WithError(err).Error("lookup")
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
		wg.Add(1)
		go doLookup(server)
	}

	done := make(chan struct{}, 1)

	go func() {
		wg.Wait()
		done <- struct{}{}
	}()

	select {
	case ret := <-result:
		return ret, nil
	case <-time.After(waitInterval):
		return nil, errors.New("query timeout")
	case <-done:
		return nil, errors.New("all lookup error")
	}
}

func questionString(q *dns.Question) string {
	return q.Name + " " + dns.TypeToString[q.Qtype]
}

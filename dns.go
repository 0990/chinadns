package chinadns

import (
	"context"
	"errors"
	"fmt"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type LookupFunc func(ctx context.Context, request *dns.Msg, server *Resolver) (reply *dns.Msg, remark string, err error)

type LookupResult struct {
	reply    *dns.Msg
	resolver *Resolver
}

func reqID(req *dns.Msg) string {
	return strconv.FormatInt(int64(req.Id), 16)
}

func (s *Server) Serve(w dns.ResponseWriter, req *dns.Msg) {
	logger := logrus.WithFields(logrus.Fields{
		"q":  questionString(&req.Question[0]),
		"id": reqID(req),
	})

	start := time.Now()

	var lookupRet *LookupResult

	var hitCache bool

	question := req.Question[0]

	reqDomain := reqDomain(req)

	defer func() {
		if !hitCache {
			s.setCached(question, lookupRet)
		}

		if lookupRet == nil {
			logger.Warn("reply==nil")
			reply := new(dns.Msg)
			reply.SetReply(req)

			lookupRet = &LookupResult{reply: reply}
		}

		var filter bool
		if attrs := s.getResolverAttr(lookupRet.resolver); len(attrs) > 0 {
			filter = filterLookupRetByAttrs(lookupRet, attrs)
		}

		// https://github.com/miekg/dns/issues/216
		lookupRet.reply.Compress = true
		_ = w.WriteMsg(lookupRet.reply)

		replyRet := replyString(lookupRet.reply)

		logger.WithFields(logrus.Fields{
			"rtt":    timeSinceMS(start),
			"dns":    lookupRet.resolver,
			"empty":  len(replyRet) == 0,
			"filter": filter,
			"result": "\n" + lookupRet.reply.String(),
			"cache":  hitCache,
		}).Debug("DNS reply")
	}()

	//自定义域名中查找
	if reply, ok := s.lookUpInCustom(reqDomain, req); ok {
		lookupRet = &LookupResult{
			reply:    reply,
			resolver: nil,
		}
		return
	}

	if v, ok := s.cache.Get(question); ok {
		hitCache = true

		r := v.(*LookupResult)
		reply := r.reply.Copy()
		reply.Id = req.Id
		lookupRet = &LookupResult{
			reply:    reply,
			resolver: r.resolver,
		}
		return
	}

	//ip反向查找域名类型查询，暂不支持
	if req.Question[0].Qtype == dns.TypePTR {
		return
	}
	//s.normalizeRequest(req)

	lookupRetChnGfw := make(chan *LookupResult)
	go func() {
		ret, err := s.lookupChnGfw(reqDomain, req, logger, start)
		if err != nil {
			lookupRetChnGfw <- nil
			logger.WithError(err).Error("query error")
			return
		}
		lookupRetChnGfw <- ret
	}()

	if len(s.DNSAdBlockServers) > 0 {
		adBlockResult, err := s.lookupAdBlock(req)
		if err == nil && adBlockResult != nil && s.DNSAdBlockJudge.IsAdBlockReply(adBlockResult.reply) {
			lookupRet = adBlockResult
			return
		} else if err != nil {
			logger.WithError(err).Error("query error")
		}
	}

	lookupRet = <-lookupRetChnGfw
}

func (s *Server) lookupChnGfw(reqDomain string, req *dns.Msg, logger *logrus.Entry, start time.Time) (*LookupResult, error) {
	//国内域名直接走国内dns
	if s.chnDomainMatcher.IsMatch(reqDomain) {
		return lookupInServers(req, s.DNSChinaServers, time.Second*2, s.lookup)
	}

	//gfw block的域名直接使用国外dns
	if s.gfwDomainMatcher.IsMatch(reqDomain) {
		return lookupInServers(req, s.DNSAbroadServers, time.Second*2, s.lookupProxyPriority)
	}

	lookupRetAbroad := make(chan *LookupResult, 1)
	go func() {
		ret, err := lookupInServers(req, s.DNSAbroadServers, time.Second*2, s.lookupProxyPriority)
		if err != nil {
			logger.WithError(err).Error("query error")
			return
		}
		lookupRetAbroad <- ret
	}()

	lookupRet, err := lookupInServers(req, s.DNSChinaServers, time.Millisecond*200, s.lookup)
	if err != nil {
		logger.WithError(err).Error("query error")
	}

	var useAbroadReason string
	if err != nil {
		useAbroadReason = "lookup china dns error"
	} else if replyRet := replyString(lookupRet.reply); replyRet == "" {
		useAbroadReason = "lookup china dns ok,but reply is empty"
	} else if !s.isReplyIPChn(lookupRet.reply) {
		useAbroadReason = "lookup china dns ok,but reply is abroad"
	}

	if useAbroadReason == "" {
		return lookupRet, err
	}

	//使用国内dns但返回的是国外ip,则用国外dns的查询结果
	logrus.WithFields(logrus.Fields{
		"domain": reqDomain,
		"reason": useAbroadReason,
	}).Warn("Try use abroad dns")

	select {
	case lookupRet = <-lookupRetAbroad:
		return lookupRet, nil
	case <-time.After(time.Second * 3):
		return nil, errors.New("lookup abroad dns timeout")
	}
}

func (s *Server) lookupAdBlock(req *dns.Msg) (*LookupResult, error) {
	return lookupInServers(req, s.DNSAdBlockServers, time.Millisecond*50, s.lookup)
}

// 查找自定义域名
func (s *Server) lookUpInCustom(domain string, req *dns.Msg) (*dns.Msg, bool) {
	ret, ok := s.Domain2IP.Load(domain)
	if !ok {
		return nil, false
	}

	logger := logrus.WithFields(logrus.Fields{
		"q":  questionString(&req.Question[0]),
		"id": reqID(req),
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
		contain, err := s.ChinaCIDR.Contains(ip)
		if err != nil {
			logrus.WithError(err).WithField("ip", ip.String()).Error("ChinaCIDR.Contains")
			return false
		}
		return contain
	}
	return true
}

func lookupInServers(req *dns.Msg, servers []*Resolver, waitInterval time.Duration, lookup LookupFunc) (*LookupResult, error) {
	if len(servers) == 0 {
		return nil, errors.New("no servers")
	}

	logger := logrus.WithFields(logrus.Fields{
		"id": reqID(req),
		"q":  questionString(&req.Question[0]),
	})

	result := make(chan *LookupResult, 1)

	//TODO miekg/dns库对context取消支持目前还不完善，这里最好是获得任一结果，取消其它查询
	ctx, cancel := context.WithTimeout(context.Background(), waitInterval)
	defer cancel()

	var wg sync.WaitGroup
	var errs MultiError
	doLookup := func(server *Resolver) {
		defer wg.Done()

		reqCopy := req.Copy()

		start := time.Now()
		reply, remark, err := lookup(ctx, reqCopy, server)
		rtt := timeSinceMS(start)
		if err != nil {
			errs.Add(fmt.Errorf("dns:%s,rtt:%v,err:%w", server, rtt, err))
			return
		}

		log := logger.WithFields(logrus.Fields{
			"rtt":    rtt,
			"dns":    server,
			"empty":  len(replyString(reply)) == 0,
			"result": "\n" + reply.String(),
		})

		if remark != "" {
			log = log.WithField("remark", remark)
		}

		log.Debug("Query result")

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
		cancel()
		return ret, nil
	case <-done:
		return nil, &errs
	}
}

func questionString(q *dns.Question) string {
	return q.Name + " " + dns.TypeToString[q.Qtype]
}

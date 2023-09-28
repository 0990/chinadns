package chinadns

import (
	"github.com/0990/chinadns/pkg/response"
	"github.com/miekg/dns"
	"time"
)

func (s *Server) setCached(question dns.Question, ret *LookupResult) {
	if ret == nil {
		return
	}

	mt, _ := response.Typify(ret.reply, time.Now().UTC())
	switch mt {
	case response.NoError, response.Delegation:
		s.cache.Set(question, ret)
	default:
	}
}

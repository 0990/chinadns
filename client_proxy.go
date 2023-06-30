package chinadns

import (
	"context"
	"github.com/0990/socks5"
	"github.com/miekg/dns"
	"time"
)

func lookUpBySocks5(ctx context.Context, proxyAddr string, network string, addr string, req *dns.Msg) (*dns.Msg, error) {
	sc := socks5.NewSocks5Client(socks5.ClientCfg{
		ServerAddr: proxyAddr,
		UserName:   "",
		Password:   "",
		UDPTimout:  60,
		TCPTimeout: 60,
	})

	timeout := time.Second * 10
	if deadline, ok := ctx.Deadline(); ok && !deadline.IsZero() {
		timeout = deadline.Sub(time.Now())
	}

	c, err := sc.DialTimeout(network, addr, timeout)
	if err != nil {
		return nil, err
	}

	defer c.Close()

	co := &dns.Conn{Conn: c} // c is your net.Conn
	err = co.WriteMsg(req)
	if err != nil {
		return nil, err
	}
	in, err := co.ReadMsg()
	if err != nil {
		return nil, err
	}
	co.Close()
	return in, nil
}

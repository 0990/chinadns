// +build go1.11
// +build aix darwin dragonfly freebsd linux netbsd openbsd

package chinadns

import (
	"context"
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

func reuseportControl(network, address string, c syscall.RawConn) error {
	var opErr error
	err := c.Control(func(fd uintptr) {
		opErr = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
		if opErr != nil {
			return
		}
		opErr = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
	})
	if err != nil {
		return err
	}

	return opErr
}

func listenTCP(network, addr string, reuseport bool) (net.Listener, error) {
	var lc net.ListenConfig
	if reuseport {
		lc.Control = reuseportControl
	}

	return lc.Listen(context.Background(), network, addr)
}

func listenUDP(network, addr string, reuseport bool) (net.PacketConn, error) {
	var lc net.ListenConfig
	if reuseport {
		lc.Control = reuseportControl
	}

	return lc.ListenPacket(context.Background(), network, addr)
}

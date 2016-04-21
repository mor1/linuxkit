// Package proxy provides a network Proxy interface and implementations for TCP
// and UDP.
package libproxy

import (
	"fmt"
	"net"
	"github.com/djs55/vsock"
)

// Proxy defines the behavior of a proxy. It forwards traffic back and forth
// between two endpoints : the frontend and the backend.
// It can be used to do software port-mapping between two addresses.
// e.g. forward all traffic between the frontend (host) 127.0.0.1:3000
// to the backend (container) at 172.17.42.108:4000.
type Proxy interface {
	// Run starts forwarding traffic back and forth between the front
	// and back-end addresses.
	Run()
	// Close stops forwarding traffic and close both ends of the Proxy.
	Close()
	// FrontendAddr returns the address on which the proxy is listening.
	FrontendAddr() net.Addr
	// BackendAddr returns the proxied address.
	BackendAddr() net.Addr
}



// NewProxy creates a Proxy according to the specified frontendAddr and backendAddr.
func NewProxy(frontendAddr *vsock.VsockAddr, backendAddr net.Addr) (Proxy, error) {
	switch backendAddr.(type) {
	case *net.UDPAddr:
		listener, err := vsock.Listen(frontendAddr.Port)
		if err != nil {
			return nil, err
		}
		return NewUDPProxy(frontendAddr, NewUDPListener(listener), backendAddr.(*net.UDPAddr))
	case *net.TCPAddr:
		listener, err := vsock.Listen(frontendAddr.Port)
		if err != nil {
			return nil, err
		}
		return NewTCPProxy(listener, backendAddr.(*net.TCPAddr))
	default:
		panic(fmt.Errorf("Unsupported protocol"))
	}
}

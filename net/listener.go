package net

import "net"

// Interface for something like net.TCPListener but not using the concrete type.
type TCPListener interface {
	net.Listener

	AcceptTCP() (DuplexConn, error)

	Close() error

	Addr() net.Addr
}

type tcpDuplexListenerAdapter struct {
	*net.TCPListener
}

func (l *tcpDuplexListenerAdapter) AcceptTCP() (DuplexConn, error) {
	return l.TCPListener.AcceptTCP()
}

func AdaptListener(l *net.TCPListener) TCPListener {
	return &tcpDuplexListenerAdapter{l}
}

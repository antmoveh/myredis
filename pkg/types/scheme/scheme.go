package scheme

import (
	"context"
	"net"
)

type ServerHandler interface {
	Handle(ctx context.Context, conn net.Conn, stopChan <-chan struct{})
	Close() error
}

type ServerConnection interface {
	Write([]byte) error
	// client should keep its subscribing channels
	SubsChannel(channel string)
	UnSubsChannel(channel string)
	SubsCount() int
	GetChannels() []string
}

type DB interface {
	Exec(client ServerConnection, args [][]byte) Reply
	AfterClientClose(c ServerConnection)
	Close()
}

type Reply interface {
	ToBytes() []byte
}

package server

import (
	"context"
	"net"
)

type Handler interface {
	Handle(ctx context.Context, conn net.Conn, stopChan <-chan struct{})
	Close() error
}

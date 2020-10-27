package server

import (
	"bufio"
	"context"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"sync"
)

type EchoHandler struct {
	activeConn sync.Map
}

type Client struct {
	Conn net.Conn
}

func (c *Client) Close() error {
	_ = c.Conn.Close()
	return nil
}

func (h *EchoHandler) Handle(ctx context.Context, conn net.Conn, stopChan <-chan struct{}) {

	client := &Client{
		Conn: conn,
	}
	h.activeConn.Store(client, 1)

	reader := bufio.NewReader(conn)
	for {
		// may occurs: client EOF, client timeout, server early close
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				logrus.Info("connection close")
				h.activeConn.Delete(conn)
			} else {
				logrus.Warn(err)
			}
			return
		}
		// time.Sleep(10 * time.Second)
		b := []byte(msg)
		_, _ = conn.Write(b)
	}
}

func (h *EchoHandler) Close() error {
	h.activeConn.Range(func(key interface{}, val interface{}) bool {
		client := key.(*Client)
		err := client.Close()
		if err != nil {
			logrus.Info("disconnect client faild: %s", err.Error())
		}
		logrus.Info("disconnect redis client...")
		return true
	})
	return nil
}

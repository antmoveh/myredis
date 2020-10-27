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

func MakeEchoHandler() *EchoHandler {
	return &EchoHandler{}
}

type Client struct {
	Conn net.Conn
}

func (c *Client) Close() error {
	// c.Waiting.WaitWithTimeout(10 * time.Second)
	c.Conn.Close()
	return nil
}

func (h *EchoHandler) Handle(ctx context.Context, conn net.Conn) {

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
		// client.Waiting.Add(1)
		//l ogger.Info("sleeping")
		// time.Sleep(10 * time.Second)
		b := []byte(msg)
		conn.Write(b)
		// client.Waiting.Done()
	}
}

func (h *EchoHandler) Close() error {
	logrus.Info("handler shuting down...")
	// TODO: concurrent wait
	h.activeConn.Range(func(key interface{}, val interface{}) bool {
		client := key.(*Client)
		client.Close()
		return true
	})
	return nil
}

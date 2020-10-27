package tcp

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"myredis/pkg/redis/server"
	"net"
	"sync"
)

type ListenerServe struct {
	Ip       string
	Port     int
	StopChan <-chan struct{}
	Handle   server.Handler
	Listener net.Listener
}

func (ls *ListenerServe) InitListenerServe(ip string, port int, stopChan <-chan struct{}, handle server.Handler) {
	ls.Ip = ip
	ls.Port = port
	ls.StopChan = stopChan
	ls.Handle = handle
}

func (ls *ListenerServe) Start() {
	address := fmt.Sprintf("%s:%d", ls.Ip, ls.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		logrus.Fatal(fmt.Sprintf("listen err: %v", err))
	}
	ls.Listener = listener
	// listen port
	logrus.Info(fmt.Sprintf("bind: %s, start listening...", address))
	defer func() {
		_ = listener.Close()
		_ = ls.Handle.Close()
	}()

	go ls.Close()

	ctx, _ := context.WithCancel(context.Background())
	var waitDone sync.WaitGroup
	for {
		conn, err := listener.Accept()
		if err != nil {
			_, isClose := <-ls.StopChan
			if !isClose {
				logrus.Info("exit redis server")
				return
			}
			logrus.Error(fmt.Sprintf("accept err: %v", err))
			continue
		}
		logrus.Info("accept new link")
		go func() {
			defer func() {
				waitDone.Done()
			}()
			waitDone.Add(1)
			ls.Handle.Handle(ctx, conn, ls.StopChan)
		}()
	}
}

func (ls *ListenerServe) Close() {
	for {
		select {
		case <-ls.StopChan:
			logrus.Info("close tcp connect")
			_ = ls.Listener.Close()
			_ = ls.Handle.Close()
			return
		default:
		}
	}
}
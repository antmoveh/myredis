package tcp

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"myredis/pkg/configuration"
	"myredis/pkg/redis/server"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type ListenerServe struct {
	Ip       string
	Port     int
	StopChan <-chan struct{}
	Handle   server.Handler
}

func (ls *ListenerServe) InitListenerServe(ip string, port int, stopChan <-chan struct{}, handle server.Handler) {
	ls.Ip = ip
	ls.Port = port
	ls.StopChan = stopChan
	ls.Handle = handle
}

func (ls *ListenerServe) Start() {
	address := fmt.Sprintf("%s:%d", configuration.BindIpAddress, configuration.BindPort)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		logrus.Fatal(fmt.Sprintf("listen err: %v", err))
	}

	// listen signal
	stopChan := make(chan struct{})
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigCh
		switch sig {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			logrus.Info("shuting down...")
			close(stopChan)
			_ = listener.Close() // listener.Accept() will return err immediately
			_ = Handle.Close()   // close connections
		}
	}()

	// listen port
	logrus.Info(fmt.Sprintf("bind: %s, start listening...", address))
	defer func() {
		_ = listener.Close()
		_ = handler.Close()
	}()

	ctx, _ := context.WithCancel(context.Background())
	var waitDone sync.WaitGroup

	for {
		conn, err := listener.Accept()
		if err != nil {
			logrus.Error(fmt.Sprintf("accept err: %v", err))
			continue
		}
		logrus.Info("accept new link")
		go func() {
			defer func() {
				waitDone.Done()
			}()
			waitDone.Add(1)
			handler.Handle(ctx, conn, stopChan)
		}()
	}
}

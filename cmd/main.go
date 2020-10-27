package main

import (
	"github.com/sirupsen/logrus"
	"myredis/pkg/configuration"
	"myredis/pkg/redis/server"
	"myredis/pkg/tcp"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	configuration.InitializeConfigurations()
	stopChan := make(chan struct{})

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		switch <-sigCh {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			close(stopChan)
			logrus.Info("shutdown...")
		}
	}()

	ls := tcp.ListenerServe{}
	ls.InitListenerServe(configuration.BindIpAddress, configuration.BindPort, stopChan, &server.EchoHandler{})
	ls.Start()


}

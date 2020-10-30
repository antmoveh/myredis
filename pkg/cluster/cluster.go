package cluster

import (
	"fmt"
	"github.com/golang/groupcache/consistenthash"
	"github.com/sirupsen/logrus"
	"myredis/pkg/configuration"
	"myredis/pkg/db"
	"myredis/pkg/redis/client"
	"myredis/pkg/redis/reply"
	"myredis/pkg/types/scheme"
	"runtime/debug"
	"strings"
)

type Cluster struct {
	self string

	db         *db.DB
	peerPicker *consistenthash.Map
	peers      map[string]*client.Client
}

const (
	replicas = 4
)

func MakeCluster() *Cluster {
	cluster := &Cluster{
		self: configuration.SelfServerAddress,

		db:         db.MakeDB(),
		peerPicker: consistenthash.New(replicas, nil),
		peers:      make(map[string]*client.Client),
	}
	if configuration.PeerServerAddress != nil && len(configuration.PeerServerAddress) > 0 && configuration.SelfServerAddress != "" {
		contains := make(map[string]bool)
		peers := make([]string, len(configuration.PeerServerAddress)+1)[:]
		for _, peer := range configuration.PeerServerAddress {
			if _, ok := contains[peer]; ok {
				continue
			}
			contains[peer] = true
			peers = append(peers, peer)
		}
		peers = append(peers, configuration.SelfServerAddress)
		cluster.peerPicker.Add(peers...)
	}
	return cluster
}

// args contains all
type CmdFunc func(cluster *Cluster, c scheme.ServerConnection, args [][]byte) scheme.Reply

func (cluster *Cluster) Close() {
	cluster.db.Close()
}

var router = MakeRouter()

func (cluster *Cluster) Exec(c scheme.ServerConnection, args [][]byte) (result scheme.Reply) {
	defer func() {
		if err := recover(); err != nil {
			logrus.Warn(fmt.Sprintf("error occurs: %v\n%s", err, string(debug.Stack())))
			result = &reply.UnknownErrReply{}
		}
	}()

	cmd := strings.ToLower(string(args[0]))
	cmdFunc, ok := router[cmd]
	if !ok {
		return reply.MakeErrReply("ERR unknown command '" + cmd + "', or not supported in cluster mode")
	}
	result = cmdFunc(cluster, c, args)
	return
}

func (cluster *Cluster) AfterClientClose(c scheme.ServerConnection) {

}

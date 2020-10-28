package db

import (
	"myredis/pkg/redis/reply"
	"myredis/pkg/redis/server"
)

type DB interface {
	Exec(client server.Connection, args [][]byte) reply.Reply
	AfterClientClose(c server.Connection)
	Close()
}

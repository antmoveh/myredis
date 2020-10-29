package db

import (
	"myredis/pkg/redis/reply"
	"myredis/pkg/types/scheme"
)

func Ping(db *DB, args [][]byte) scheme.Reply {
	if len(args) == 0 {
		return &reply.PongReply{}
	} else if len(args) == 1 {
		return reply.MakeStatusReply("\"" + string(args[0]) + "\"")
	} else {
		return reply.MakeErrReply("ERR wrong number of arguments for 'ping' command")
	}
}

package server

import (
	"bufio"
	"context"
	"github.com/sirupsen/logrus"
	"io"
	"myredis/pkg/db"
	"myredis/pkg/redis/reply"
	"myredis/pkg/types/scheme"
	"net"
	"strconv"
	"strings"
	"sync"
)

var (
	UnknownErrReplyBytes = []byte("-ERR unknown\r\n")
)

type Handler struct {
	activeConn sync.Map // *client -> placeholder
	db         scheme.DB
}

func MakeHandler() *Handler {
	var db2 scheme.DB
	// if config.Properties.Peers != nil &&
	// 	len(config.Properties.Peers) > 0 {
	// 	db = cluster.MakeCluster()
	// } else {
	// 	db = DBImpl.MakeDB()
	// }
	db2 = db.MakeDB()
	return &Handler{
		db: db2,
	}
}

func (h *Handler) closeClient(client *Client) {
	_ = client.Close()
	h.db.AfterClientClose(client)
	h.activeConn.Delete(client)
}

func (h *Handler) Handle(ctx context.Context, conn net.Conn, stopChan <-chan struct{}) {

	_, isClose := <-stopChan
	if !isClose {
		_ = conn.Close()
		return
	}

	client := MakeClient(conn)
	h.activeConn.Store(client, 1)

	reader := bufio.NewReader(conn)
	var fixedLen int64 = 0
	var err error
	var msg []byte
	// 解析RESP协议
	// 解析文本协议
	for {
		if fixedLen == 0 {
			msg, err = reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF || err == io.ErrUnexpectedEOF || strings.Contains(err.Error(), "use of closed network connection") {
					logrus.Info("connection close")
				} else {
					logrus.Warn(err)
				}
				// after client close
				h.closeClient(client)
				return // io error, disconnect with client
			}
			if len(msg) == 0 || msg[len(msg)-2] != '\r' {
				errReply := &reply.ProtocolErrReply{Msg: "invalid multibulk length"}
				_, _ = client.conn.Write(errReply.ToBytes())
			}
		} else {
			msg = make([]byte, fixedLen+2)
			_, err = io.ReadFull(reader, msg)
			if err != nil {
				if err == io.EOF || err == io.ErrUnexpectedEOF || strings.Contains(err.Error(), "use of closed network connection") {
					logrus.Info("connection close")
				} else {
					logrus.Warn(err)
				}
				// after client close
				h.closeClient(client)
				return // io error, disconnect with client
			}
			if len(msg) == 0 || msg[len(msg)-2] != '\r' || msg[len(msg)-1] != '\n' {
				errReply := &reply.ProtocolErrReply{Msg: "invalid multibulk length"}
				_, _ = client.conn.Write(errReply.ToBytes())
			}
			fixedLen = 0
		}

		if !client.uploading {
			// new request
			if msg[0] == '*' {
				// bulk multi msg
				expectedLine, err := strconv.ParseUint(string(msg[1:len(msg)-2]), 10, 32)
				if err != nil {
					_, _ = client.conn.Write(UnknownErrReplyBytes)
					continue
				}
				client.uploading = true
				client.expectedArgsCount = uint32(expectedLine)
				client.receivedCount = 0
				client.args = make([][]byte, expectedLine)
			} else {
				// text protocol
				// remove \r or \n or \r\n in the end of line
				str := strings.TrimSuffix(string(msg), "\n")
				str = strings.TrimSuffix(str, "\r")
				strs := strings.Split(str, " ")
				args := make([][]byte, len(strs))
				for i, s := range strs {
					args[i] = []byte(s)
				}

				// send reply
				result := h.db.Exec(client, args)
				if result != nil {
					_ = client.Write(result.ToBytes())
				} else {
					_ = client.Write(UnknownErrReplyBytes)
				}
			}
		} else {
			// receive following part of a request
			line := msg[0 : len(msg)-2]
			if line[0] == '$' {
				fixedLen, err = strconv.ParseInt(string(line[1:]), 10, 64)
				if err != nil {
					errReply := &reply.ProtocolErrReply{Msg: err.Error()}
					_, _ = client.conn.Write(errReply.ToBytes())
				}
				if fixedLen <= 0 {
					errReply := &reply.ProtocolErrReply{Msg: "invalid multibulk length"}
					_, _ = client.conn.Write(errReply.ToBytes())
				}
			} else {
				client.args[client.receivedCount] = line
				client.receivedCount++
			}

			// if sending finished
			if client.receivedCount == client.expectedArgsCount {
				client.uploading = false // finish sending progress

				// send reply
				result := h.db.Exec(client, client.args)
				if result != nil {
					_ = client.Write(result.ToBytes())
				} else {
					_ = client.Write(UnknownErrReplyBytes)
				}

				// finish reply
				client.expectedArgsCount = 0
				client.receivedCount = 0
				client.args = nil
			}
		}
	}
}

func (h *Handler) Close() error {
	h.activeConn.Range(func(key interface{}, val interface{}) bool {
		client := key.(*Client)
		err := client.Close()
		if err != nil {
			logrus.Info("disconnect client failed: %s", err.Error())
		}
		logrus.Info("disconnect redis client...")
		return true
	})
	h.db.Close()
	return nil
}

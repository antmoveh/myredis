package pubsub

import (
    "myredis/pkg/types/data/dict"
    "myredis/pkg/types/data/lock"
)

type Hub struct {
    // channel -> list(*Client)
    subs dict.Dict
    // lock channel
    subsLocker *lock.Locks
}

func MakeHub() *Hub {
    return &Hub{
        subs:       dict.MakeConcurrent(4),
        subsLocker: lock.Make(16),
    }
}

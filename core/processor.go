package qlimiter

import (
    "github.com/andhm/qlimiter-go/protocol"
    "github.com/andhm/qlimiter-go/util"
)

const (
    defaultQueueCap     = 10240
)

type Queue struct {
    info    chan queueInfo
}

type queueInfo struct {
    channel     *Channel
    recvMsg     *protocol.Message
}

func NewQueue(opts *util.Options) *Queue {
    var queueCap uint
    if opts.ChannelQueueCap <= 0 {
        queueCap = defaultQueueCap
    } else {
        queueCap = opts.ChannelQueueCap
    }

    q := &Queue {info : make(chan queueInfo, queueCap)}

    return q
}

func (q *Queue) Push(channel *Channel, recvMsg *protocol.Message) error {
    // logger.Info("Ready to push. RequestId[%v] ClientIp[%s]", recvMsg.RequestId(), channel.ClientIp)
    logger.Info("Push... %s", channel.ClientIp)
    qInfo := queueInfo {channel: channel, recvMsg: recvMsg}
    q.info <- qInfo
    return nil
}

func (q *Queue) Process(qlimiter *Qlimiter) error {
    for {
        qInfo :=<- q.info
        res, val, err := qlimiter.Limit(qInfo.recvMsg)
        errmsg := "ok"
        if err != nil {
            errmsg = err.Error()
        }
        // res , val := 1, 2
        buf := protocol.EncodeForResponse(qInfo.recvMsg, res, int(val), errmsg)
        qInfo.channel.Send(buf.Bytes())
    }
    return nil
}

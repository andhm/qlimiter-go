package qlimiter

import (
    "time"
    "fmt"
    "bufio"
    "net"
    "io"
    "sync"

    "github.com/andhm/qlimiter-go/protocol"
    "github.com/andhm/qlimiter-go/util"
)

var (
    ErrChannelShutdown         = fmt.Errorf("The channel has been shutdown")
    ErrRequestTimeout          = fmt.Errorf("Request timeout")
)

type Channel struct {
    // connection
    conn            io.ReadWriteCloser
    bufRead         *bufio.Reader
    ClientIp        string

    // recv
    recvChan        chan recvData

    // send
    sendChan        chan sendData

    // shutdown
    shutdown        bool
    shutdownChan    chan struct{}
    shutdownLock    sync.Mutex
}

type sendData struct {
    data    []byte
}

type recvData struct {
    msg     *protocol.Message
}

func (channel *Channel) Recv(timeOut time.Duration) (*protocol.Message, error) {
    timer := time.NewTimer(timeOut)
    defer timer.Stop()

    select {
    case recvData, ok := <- channel.recvChan:
        if !ok {
            logger.Debug("Recv recvChan shutdown. ClientIp[%s]", channel.ClientIp)
            return nil, ErrChannelShutdown
        }
        recvMsg := recvData.msg
        logger.Debug("Recv succ. RequestId[%v], ClientIp[%s]", recvMsg.RequestId(), channel.ClientIp)
        return recvMsg, nil
    case <-timer.C:
        logger.Warn("Recv timeout. ClientIp[%s]", channel.ClientIp)
        // close
        channel.close()
        return nil, ErrRequestTimeout
    case <-channel.shutdownChan:
        logger.Debug("Recv channel shutdown. ClientIp[%s]", channel.ClientIp)
        return nil, ErrChannelShutdown
    }
}

func (channel *Channel) Send(sendMsg []byte) error {
    defer func() {
        // some error that sending to closed channel, unusually
        if err:=recover(); err!=nil {
            logger.Warn("Send panic - %v", err)
        }
    }()

    sendData := sendData {data: sendMsg}
    select {
    case channel.sendChan <- sendData:
        logger.Debug("Send succ. ClientIp[%s]", channel.ClientIp)
        return nil
    case <-channel.shutdownChan:
        logger.Debug("Send channel shutdown. ClientIp[%s]", channel.ClientIp)
        return ErrChannelShutdown
    }
}

func (channel *Channel) recv () {
    if err := channel.recvLoop(); err != nil {
        if err.Error() != "EOF" {
            logger.Warn("recv error - %v. ClientIp[%s]", err, channel.ClientIp)
        } else {
            logger.Debug("recv error - %v. ClientIp[%s]", err, channel.ClientIp)
        }
        channel.close()
        return
    }
}

func (channel *Channel) recvLoop() error {
    // close recv channel
    defer channel.closeRecvChan()

    for {
        if channel.shutdown {
            logger.Debug("recvLoop channel shutdown. ClientIp[%s]", channel.ClientIp)
            return ErrChannelShutdown
        }
        msg, err := protocol.DecodeFromReader(channel.bufRead)
        if err != nil {
            return err
        }
        recvData := recvData {msg : msg}
        channel.recvChan <- recvData
    }
}

func (channel *Channel) send() {
    for {
        select {
        case sendData, ok :=<- channel.sendChan:
            if !ok {
                logger.Debug("send sendChan shutdown. ClientIp[%s]", channel.ClientIp)
                return
            }
            if sendData.data != nil {
                sent := 0
                for sent < len(sendData.data) {
                    n, err := channel.conn.Write(sendData.data[sent:])
                    if err != nil {
                        logger.Warn("send error - %v. ClientIp[%s]", err, channel.ClientIp)
                        channel.close()
                        // close send channel
                        channel.closeSendChan()
                        return
                    }
                    sent += n
                }
            }
        case <- channel.shutdownChan:
            logger.Debug("send channel shutdown. ClientIp[%s]", channel.ClientIp)
            // close send channel
            channel.closeSendChan()
            return
        }
    }
}

func (channel *Channel) close() error {
    channel.shutdownLock.Lock()
    defer channel.shutdownLock.Unlock()

    if channel.shutdown {
        return nil
    }
    logger.Debug("close succ. ClientIp[%s]", channel.ClientIp)
    channel.shutdown = true
    close(channel.shutdownChan)
    channel.conn.Close()
    return nil
}

func (channel *Channel) closeRecvChan() error {
    close(channel.recvChan)
    return nil
}

func (channel *Channel) closeSendChan() error {
    close(channel.sendChan)
    return nil
}

func BuildChannel(conn net.Conn, opts *util.Options) *Channel {
    logger.Debug("ready to build channel. ClientIp[%s]", conn.RemoteAddr().String())
    channel := &Channel {
        conn:       conn,
        bufRead:    bufio.NewReader(conn),
        ClientIp:   conn.RemoteAddr().String(),
        recvChan:   make(chan recvData, 1),
        sendChan:   make(chan sendData, 1),
        shutdown:   false,
        shutdownChan:   make(chan struct{}),
    }

    go channel.recv()
    go channel.send()

    return channel
}

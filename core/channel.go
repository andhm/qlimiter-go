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
    ErrChannelShutdown         = fmt.Errorf("The channel has been shutdown.")
    ErrRequestTimeout          = fmt.Errorf("Request timeout.")
)

const (
    defaultSendTimeout      = 1*time.Second
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
    case recvData := <- channel.recvChan:
        recvMsg := recvData.msg
        logger.Debug("Recv succ. RequestId[%v], ClientIp[%s]", recvMsg.RequestId(), channel.ClientIp)
        return recvMsg, nil
    case <-timer.C:
        logger.Warn("Recv timeout. ClientIp[%s]", channel.ClientIp)
        return nil, ErrRequestTimeout
    case <-channel.shutdownChan:
        logger.Debug("Recv channel shutdown. ClientIp[%s]", channel.ClientIp)
        return nil, ErrChannelShutdown
    }
}

func (channel *Channel) Send(sendMsg []byte, timeOut time.Duration) error {
    timer := time.NewTimer(timeOut)
    defer timer.Stop()

    sendData := sendData {data: sendMsg}
    select {
    case channel.sendChan<-sendData:
        logger.Debug("Send succ. ClientIp[%s]", channel.ClientIp)
        return nil
    case <-timer.C:
        logger.Warn("Send timeout. ClientIp[%s]", channel.ClientIp)
        return ErrRequestTimeout
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
    i := 0
    for {
        i++
        msg, err := protocol.DecodeFromReader(channel.bufRead)
        if err != nil {
            return err
        }
        recvData := recvData {msg : msg}
        channel.recvChan <- recvData
    }
}

func (channel *Channel) send() {
    timer := time.NewTimer(defaultSendTimeout)

    for {
        if !timer.Stop() {
            select {
            case <- timer.C:
            default:
            }
        }
        timer.Reset(defaultSendTimeout)

        select {
        case sendData :=<- channel.sendChan:
            if sendData.data != nil {
                sent := 0
                for sent < len(sendData.data) {
                    n, err := channel.conn.Write(sendData.data[sent:])
                    if err != nil {
                        logger.Warn("send error - %v. ClientIp[%s]", err, channel.ClientIp)
                        channel.close()
                        return
                    }
                    sent += n
                }
            }
        case <-timer.C:
            logger.Warn("send timeout. ClientIp[%s]", channel.ClientIp)
            channel.close()
            return
        case <- channel.shutdownChan:
            logger.Debug("send channel shutdown. ClientIp[%s]", channel.ClientIp)
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

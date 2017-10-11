package main

import (
    "fmt"
    "net"
    // "net/http"
    "strconv"
    // "runtime"
    "os"
    "time"
    // _ "net/http/pprof"

    "github.com/andhm/qlimiter-go/util"
    core "github.com/andhm/qlimiter-go/core"
    "github.com/andhm/qlimiter-go/log"
)

const (
    defaultPort     = 9091
)

var (
    options *util.Options
    qlimiter *core.Qlimiter
    processor *core.Queue

    logger *lg.LgWarpper
)

func main() {
    options = util.NewOptions()
    options.Parse()

    logLevel, _ := lg.ParseLogLevel(options.LogLevel, false)
    lgw, err := lg.NewLgWarpper(options.LogDir, logLevel)
    if err != nil {
        fmt.Println("Error:", err)
        os.Exit(1)
    }
    options.Lgw = lgw
    logger = lgw

    port := defaultPort
    if options.Port > 0 {
        port = options.Port
    }

    startServer(strconv.Itoa(port))
}

func startServer(addr string) {
    listen, err := net.Listen("tcp", ":"+addr)
    if err != nil {
        logger.Fatal("listen port fail. %s - %v", addr, err)
        os.Exit(1)
    }

    fmt.Printf("Qlimiter is started. port:%s\n", addr)

    qlimiter = core.New(options)
    processor = core.NewQueue(options)
    
    logger.Info("Qlimiter is started. port:%s, options:%v", addr, options)

    go processor.Process(qlimiter)

    for {
        conn, err := listen.Accept()
        if err != nil {
            logger.Error("accept fail. %s - %v", addr, err)
            os.Exit(1)
        } else {
            logger.Debug("accept success. conn:%s", conn.RemoteAddr().String())
            go qlimiterHandler(conn)
        }
    }
}

func qlimiterHandler(conn net.Conn) {
    channel := core.BuildChannel(conn, options)
    for {
        requestMsg, err := channel.Recv(1*time.Second)
        if err != nil {
            break
        }
        processor.Push(channel, requestMsg)
    }
}

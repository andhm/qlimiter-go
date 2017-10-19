package main

import (
    "fmt"
    "net"
    "strconv"
    "os"
    "time"

    // for debug
    "net/http"
    _ "net/http/pprof"

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

    if logLevel == lg.DEBUG {
        // for pprof debug
        debugServerAddress := options.DebugHttpServeIp + ":" + strconv.Itoa(options.DebugHttpServePort)
        startDebugServer(debugServerAddress)
    }

    startServer(strconv.Itoa(port))
}

func startServer(addr string) {
    listen, err := net.Listen("tcp", ":"+addr)
    if err != nil {
        fmt.Printf("Qlimiter start fail. %v\n", err)
        logger.Fatal("Listen port fail. %s - %v", addr, err)
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
            logger.Error("Accept fail. %s - %v", addr, err)
            os.Exit(1)
        } else {
            logger.Debug("Accept success. conn:%s", conn.RemoteAddr().String())
            go qlimiterHandler(conn)
        }
    }
}

func qlimiterHandler(conn net.Conn) {
    channel := core.BuildChannel(conn, options)
    for {
        requestMsg, err := channel.Recv(time.Duration(options.ProcessTimeout) * time.Millisecond)
        if err != nil {
            break
        }
        processor.Push(channel, requestMsg)
    }
}

func startDebugServer(addr string) {
    go func() {
        time.Sleep(300*time.Millisecond) // wait for Qlimiter started
        logger.Info("Ready to start Debug-Http-Server. address:%s", addr)
        err := http.ListenAndServe(addr, nil)
        if err != nil {
            logger.Info("Start Debug-Http-Server fail. %s - %v", addr, err)
        }
    }()
}

package util

import (
    "flag"
    "fmt"

    "github.com/andhm/qlimiter-go/log"
)

type Options struct {
    LogLevel        string
    LogDir          string

    Lgw             *lg.LgWarpper   // not cfg really

    Port            int
    SepNum          uint16

    ChannelQueueCap uint
    ProcessTimeout  uint

    DebugHttpServeIp    string
    DebugHttpServePort  int

}

func NewOptions() *Options {
    return &Options {}
}

func (opts *Options) Parse() error {
    port            := flag.Int("port", 9091, "listen port")
    sepNum          := flag.Int("sepnum", 10, "separate-number for 1s")
    channelQueueCap := flag.Int("channelqueuecap", 10240, "queue capacity of channel")

    logDir          := flag.String("logdir", "", "directory for log file")
    logLevel        := flag.String("loglevel", "info", "log level, e.g. debug info warn error fatal ")
    processTimeout  := flag.Int("processtimeout", 1000, "millisecs of process single request")

    debugHttpServeIp    := flag.String("debughttpserveip", "127.0.0.1", "http server ip when loglevel=debug")
    debugHttpServePort  := flag.Int("debughttpserveport", 6060, "http server port when loglevel=debug")

    flag.Parse()

    // TODO valid opts
    opts.Port               = *port
    opts.SepNum             = uint16(*sepNum)
    opts.ChannelQueueCap    = uint(*channelQueueCap)
    opts.LogDir             = *logDir
    opts.LogLevel           = *logLevel
    opts.ProcessTimeout     = uint(*processTimeout)
    opts.DebugHttpServeIp   = *debugHttpServeIp
    opts.DebugHttpServePort = *debugHttpServePort

    return nil
}

func (opts *Options) String() string {
    return fmt.Sprintf("port[%d], loglevel[%s], logdir[%s], sepnum[%v], channelqueuecap[%v], processtimeout[%v], debughttpserveip[%s], debughttpserveport[%d]", 
        opts.Port, 
        opts.LogLevel, 
        opts.LogDir, 
        opts.SepNum, 
        opts.ChannelQueueCap,
        opts.ProcessTimeout,
        opts.DebugHttpServeIp,
        opts.DebugHttpServePort)
}

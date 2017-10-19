package lg

import (
    "fmt"
    "os"
    "io"
    "sync"
    "time"
    "strconv"
    "syscall"
)

const (
    logFilePrefix       = "qlimiter"
    logInfoSeparator    = " - "
)

type LgWarpper struct {
    cfgLevel    LogLevel

    lock    sync.RWMutex
    f       *os.File
    logDir  string

    today   int

    logger  Logger
}

type lgWriter struct {
    lgw     *LgWarpper 
}

func (lw lgWriter) Output(maxdepth int, s string) error {
    now := time.Now()
    today := now.Format("20060102")
    datetime := now.Format("2006-01-02 15:04:05")
    todayInt, _ := strconv.Atoi(today)

    lgw := lw.lgw
    for {
        lgw.lock.RLock()
        if lgw.today == todayInt {
            lgw.lock.RUnlock()
            break
        } else {
            lgw.lock.RUnlock()
            lgw.lock.Lock()
            if lgw.today == todayInt {
                lgw.lock.Unlock()
                break
            }
            logFileName := logFilePrefix + ".log." + today
            logFilePath := lgw.logDir + "/" + logFileName
            f, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
            if err == nil {
                oldF := lgw.f

                lgw.f = f
                lgw.today = todayInt
                // os.Stderr = f
                // os.Stdout = f
                syscall.Dup2(int(f.Fd()), 1)
                syscall.Dup2(int(f.Fd()), 2)

                // close old file
                time.AfterFunc(10*time.Second, func() {
                    oldF.Close()
                })
            }
            lgw.lock.Unlock()
        }
        break
    }
    _, err := io.WriteString(lgw.f, datetime + logInfoSeparator + s)
    return err
}

func NewLgWarpper(logDir string, lvl LogLevel) (*LgWarpper, error) {
    lgw := &LgWarpper {cfgLevel : lvl}

    if logDir == "" {
        workDir, _ := os.Getwd()
        logDir = workDir + "/logs"
        if _, err := os.Stat(logDir); os.IsNotExist(err) {
            if err = os.Mkdir(logDir, os.ModePerm); err != nil {
                return nil, fmt.Errorf("Create log dir [%s] fail", logDir)
            }
        }
    }  else if _, err := os.Stat(logDir); os.IsNotExist(err) {
        return nil, fmt.Errorf("Log-Directory [%s] not found", logDir)
    }

    lgw.logDir = logDir

    logger := lgWriter {lgw : lgw}

    lgw.logger = logger

    return lgw, nil
}

func (lgw *LgWarpper) Logf(msgLevel LogLevel, f string, args ...interface{}) {
    if lgw.cfgLevel > msgLevel {
        return
    }

    lgw.logger.Output(3, fmt.Sprintf(msgLevel.String()+logInfoSeparator+f, args...))
}



func (lgw *LgWarpper) Debug(f string, args ...interface{}) {
    lgw.Logf(DEBUG, f+"\n", args...)
}

func (lgw *LgWarpper) Info(f string, args ...interface{}) {
    lgw.Logf(INFO, f+"\n", args...)
}

func (lgw *LgWarpper) Warn(f string, args ...interface{}) {
    lgw.Logf(WARN, f+"\n", args...)
}

func (lgw *LgWarpper) Error(f string, args ...interface{}) {
    lgw.Logf(ERROR, f+"\n", args...)
}

func (lgw *LgWarpper) Fatal(f string, args ...interface{}) {
    lgw.Logf(FATAL, f+"\n", args...)
}

package qlimiter

import (
    // "sync"
    "time"
    "strconv"

    "github.com/andhm/qlimiter-go/protocol"
    "github.com/andhm/qlimiter-go/util"
    "github.com/andhm/qlimiter-go/log"
)

var logger *lg.LgWarpper

const (
    defaultSepNum       = 10
    maxSepNum           = 1000
)

const (
    RES_SUCC    = int(1)
    RES_FAIL    = int(0)

    RES_ERR     = int(-1)
)

type Qlimiter struct {
    qpsMap      map[string]*qps
    // lock        sync.RWMutex
    sepNum      uint16
}

type qps struct {
    currIdx     uint8
    qpsInfo     [2]*qpsInfo
}

type qpsInfo struct {
    sepVals     []uint
    time        int64
}

func New(opts *util.Options) *Qlimiter {
    q := &Qlimiter {
        qpsMap:     make(map[string]*qps),
    }

    if opts.SepNum <= 0 {
        q.sepNum = defaultSepNum
    } else if opts.SepNum > maxSepNum {
        q.sepNum = maxSepNum
    } else {
        q.sepNum = opts.SepNum
    }

    logger = opts.Lgw

    return q
}

func (qlimiter *Qlimiter) Limit(request *protocol.Message, clientIp string) (int, int64, error) {
    logger.Info("Limit interface. RequestId[%v] ClientIp[%s]", request.RequestId(), clientIp)
    // TODO valid params
    meta        := request.Meta

    key         := meta[protocol.M_Key]
    maxVal, _   := strconv.ParseInt(meta[protocol.M_Maxval], 10, 64)
    initVal, _  := strconv.ParseInt(meta[protocol.M_Initval], 10, 64)
    step, _     := strconv.ParseInt(meta[protocol.M_Step], 10, 64)

    return qlimiter.limit(key, maxVal, initVal, step)
}

func (qlimiter *Qlimiter) limit(key string, maxVal int64, initVal int64, step int64) (int, int64, error) {
    nowUnixNano := time.Now().UnixNano()
    timeSec := nowUnixNano/1e9
    timeMsec := (nowUnixNano/1e6)%1e3
    idx := qlimiter.buildSepIdx(timeMsec)

    logger.Debug("timeSec [%d], timeMsec [%d] idx [%d]", timeSec, timeMsec, idx)

    // qlimiter.lock.RLock()
    theQps, ok := qlimiter.qpsMap[key]
    if !ok {
        // qlimiter.lock.RUnlock()
        // qlimiter.lock.Lock()
        logger.Debug("not found key [%s]", key)
        theQpsInfo := &qpsInfo {sepVals : make([]uint, qlimiter.sepNum), time : timeSec}
        theQpsInfo.sepVals[idx] = uint(step)

        theQps = &qps {currIdx : 0}
        theQps.qpsInfo[0] = theQpsInfo
        theQps.qpsInfo[1] = &qpsInfo {sepVals : make([]uint, qlimiter.sepNum), time : int64(0)}

        qlimiter.qpsMap[key] = theQps

        // qlimiter.lock.Unlock()
        return RES_SUCC, step, nil
    }

    // qlimiter.lock.RUnlock()
    // qlimiter.lock.Lock()
    // defer qlimiter.lock.Unlock()

    currQpsInfo := theQps.qpsInfo[theQps.currIdx]
    prevIdx := uint8(0)
    if theQps.currIdx == 0 {
        prevIdx = 1
    }

    prevQpsInfo := theQps.qpsInfo[prevIdx]

    if currQpsInfo.time == timeSec {
        logger.Debug("same time, update it")
        i := int16(idx)
        j := uint16(0)
        currVal := int64(0)
        for ; j < qlimiter.sepNum; j++ {
            _idx := uint16(0)
            if i < 0 {
                _idx = uint16(int16(qlimiter.sepNum) + i)
                currVal += int64(prevQpsInfo.sepVals[_idx])
            } else {
                _idx = uint16(i)
                currVal += int64(currQpsInfo.sepVals[_idx])
            }
            if currVal + step > maxVal {
                return RES_FAIL, currVal, nil
            }
            i--
        }

        currQpsInfo.sepVals[idx] += uint(step)
        return RES_SUCC, currVal + step, nil

    } else if timeSec - currQpsInfo.time == 1 {
        logger.Debug("next time, swap it")
        if theQps.currIdx == 0 {
            theQps.currIdx = 1
            prevIdx = 0
        } else {
            theQps.currIdx = 0
            prevIdx = 1
        }

        currQpsInfo = theQps.qpsInfo[theQps.currIdx]
        prevQpsInfo = theQps.qpsInfo[prevIdx]

        resetSepVals(currQpsInfo.sepVals)
        currQpsInfo.time = timeSec
        i := uint(qlimiter.sepNum) - 1
        currVal := int64(0)
        for ; i > idx; i-- {
            currVal += int64(prevQpsInfo.sepVals[i])
            if currVal + step > maxVal {
                return RES_FAIL, currVal, nil
            }
        }

        currQpsInfo.sepVals[idx] = uint(step)
        return RES_SUCC, currVal + step, nil

    } else {
        logger.Debug("not same time, reset it")
        resetSepVals(currQpsInfo.sepVals)
        resetSepVals(prevQpsInfo.sepVals)

        currQpsInfo.time = timeSec
        prevQpsInfo.time = timeSec
        currQpsInfo.sepVals[idx] = uint(step)
        return RES_SUCC, step, nil
    }
}

func (qlimiter *Qlimiter) buildSepIdx(msec int64) uint {
    var div uint
    if (qlimiter.sepNum <= 10) {
        div = 100
    } else if (qlimiter.sepNum <= 100) {
        div = 10
    } else if (qlimiter.sepNum <= 1000) {
        div = 1
    }
    idx := (uint(msec)/div)%uint(qlimiter.sepNum)
    return idx
}

func resetSepVals(sepVals []uint) {
    for i, _ := range sepVals {
        sepVals[i] = 0
    }
}


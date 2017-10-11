package util

import (
    "io"
    "bufio"
    "encoding/binary"
)

func ReadInt32(buf io.Reader) int32 {
    var i int32
    binary.Read(buf, binary.BigEndian, &i)
    return i
}

func ReadBytes(buf *bufio.Reader, size int) ([]byte, error) {
    tempbytes := make([]byte, size)
    var s, n int = 0, 0
    var err error
    for s < size && err == nil {
        n, err = buf.Read(tempbytes[s:])
        s += n
    }
    return tempbytes, err
}
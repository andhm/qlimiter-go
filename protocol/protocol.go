package protocol

import (
    "bufio"
    "bytes"
    "encoding/binary"
    "strings"
    // "fmt"
    "strconv"

    "github.com/andhm/qlimiter-go/util"
)

const (
    RequestType     = iota
    ResponseType
)

const (
    QlimiterMagic   = 0xf2f2
    QlimiterVersion = 0x01
)

const (
    MethodQps      = iota
    MethodIncr
    MethodDecr
)

const (
    M_Method        = "M_t"
    M_Key           = "M_k"
    M_Maxval        = "M_m"
    M_Step          = "M_s"
    M_Initval       = "M_i"

    M_Result        = "M_r"
    M_Currval       = "M_c"
    M_Error         = "M_e"
)

type Header struct {
    Magic       uint16
    Version     uint8
    RequestId   uint64
}

type Message struct {
    Header      *Header
    Meta        map[string]string
    Type        int
}

func BuildHeader(requestId uint64, version uint8) *Header {
    header := &Header{QlimiterMagic, version, requestId}
    return header;
}

func BuildResponseHeader(requestId uint64) *Header {
    return BuildHeader(requestId, QlimiterVersion)
}

func (msg *Message) RequestId() uint64 {
    return msg.Header.RequestId
}

func (msg *Message) Encode() (buf *bytes.Buffer) {
    buf = new(bytes.Buffer)
    var metaBuf bytes.Buffer
    meta := msg.Meta
    for k, v := range meta {
        metaBuf.WriteString(k)
        metaBuf.WriteString("\n")
        metaBuf.WriteString(v)
        metaBuf.WriteString("\n")
    }
    var metaSize int32
    if metaBuf.Len() > 0 {
        metaSize = int32(metaBuf.Len() - 1)
    } else {
        metaSize = 0
    }

    binary.Write(buf, binary.BigEndian, msg.Header)
    binary.Write(buf, binary.BigEndian, metaSize)
    if metaSize > 0 {
        binary.Write(buf, binary.BigEndian, metaBuf.Bytes()[:metaSize])
    }
    return buf
}

func EncodeForResponse(msg *Message, res int, curr int, errmsg string) (buf *bytes.Buffer) {
    responseMsg := &Message {
        Header : BuildResponseHeader(msg.Header.RequestId), 
        Meta : map[string]string {
            M_Result : strconv.Itoa(res), 
            M_Currval : strconv.Itoa(curr),
            M_Error : errmsg,
        }, 
        Type : ResponseType,
    }

    buf = responseMsg.Encode()
    return buf
}

func DecodeFromReader(buf *bufio.Reader) (msg *Message, err error) {
    header := &Header{}
    err = binary.Read(buf, binary.BigEndian, header)
    if err != nil {
        return nil, err
    }

    metaSize := util.ReadInt32(buf)
    metaMap := make(map[string]string)
    if metaSize > 0 {
        meta, err := util.ReadBytes(buf, int(metaSize))
        if (err != nil) {
            return nil, err
        }

        values := strings.Split(string(meta), "\n")

        for i := 0; i < len(values); i++ {
            key := values[i]
            i++
            metaMap[key] = values[i]
        } 
    }

    msg = &Message{header, metaMap, RequestType}
    return msg, err
}

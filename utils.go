package transport

import (
	"go.uber.org/zap"
	"io"
	"math/rand"
	"net"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano()) // init rand
}

const (
	BufferLength = 8 * 1024
	BufferLimit  = BufferLength - 512	// 保留一部分用作 header
)

var (
	bufferPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, BufferLength)
		}}
)

// getRandomBytes 生成一个随机字节
func getRandomBytes() byte {
	bin := []byte{0x00}
	rand.Read(bin)
	return bin[0]
}

func GetBuffer() []byte {
	return bufferPool.Get().([]byte)
}

func PutBuffer(buffer []byte) {
	bufferPool.Put(buffer)
}

func Relay(reader, writer io.ReadWriteCloser) {
	buffer := GetBuffer()
	defer PutBuffer(buffer)

	for {
		readerConn, ok := reader.(net.Conn)
		if ok {
			err := readerConn.SetReadDeadline(time.Now().Add(8 * time.Second)) // 默认读取超时时间为八秒
			if err != nil {
				zap.L().Debug("reader conn set read deadline err", zap.Error(err))
				return
			}
			reader = readerConn
		}

		n, err := reader.Read(buffer[:BufferLimit])
		if err != nil {
			zap.L().Debug("err on read reader msg", zap.Error(err))
			return
		}

		_, err = writer.Write(buffer[:n])
		if err != nil {
			zap.L().Debug("err on write writer msg", zap.Error(err))
			return
		}
	}
}

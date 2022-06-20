package transport

import (
	"github.com/Jack-Kingdom/go-dsa/buffer"
	"math/rand"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano()) // init rand
}

const (
	BufferLength = 8 * 1024
	DataLimit    = 7 * 1024 // 保留一部分用作 下层协议传输的 header
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
	return buffer.Get(BufferLength)
}

func PutBuffer(buffer []byte) {
	bufferPool.Put(buffer)
}

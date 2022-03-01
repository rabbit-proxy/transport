package transport

import (
	"math/rand"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano()) // init rand
}

const (
	BufferLength = 16 * 1024
	BufferLimit  = 15 * 1024 // 保留一部分用作 header
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
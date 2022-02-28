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
	DefaultBufferLength = 8 * 1024
)

var (
	bufferLength = 0

	bufferPool   = sync.Pool{
		New: func() interface{} {
			if bufferLength == 0 {
				return make([]byte, DefaultBufferLength)
			}
			return make([]byte, bufferLength)
		}}
)

// getRandomBytes 生成一个随机字节
func getRandomBytes() byte {
	bin := []byte{0x00}
	rand.Read(bin)
	return bin[0]
}

// InitBufferLength 初始化 buffer 长度，仅在获取 buffer 前调用
func InitBufferLength(num int) {
	bufferLength = num
}

func GetBuffer() []byte {
	return bufferPool.Get().([]byte)
}

func PutBuffer(buffer []byte) {
	bufferPool.Put(buffer)
}

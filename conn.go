package transport

import (
	"github.com/gorilla/websocket"
	"net"
)

// Socket 基础的连接接口
type Socket interface {
	Write(data []byte) (int, error)
	Read(data []byte) (int, error)
	Close() error
}

// TCPSocketType 基于 tcp 的传输方式
type TCPSocketType struct {
	Encryption
	Conn *net.TCPConn
}

func (socket *TCPSocketType) Write(data []byte) (int, error) {
	socket.Encrypt(data)
	return socket.Conn.Write(data)
}

func (socket *TCPSocketType) Read(data []byte) (int, error) {
	n, err := socket.Conn.Read(data)
	if err != nil {
		return n, err
	}
	socket.Decrypt(data[:n])
	return n, nil
}

func (socket *TCPSocketType) Close() error {
	return socket.Conn.Close()
}

// WSType 基于 websocket 的传输方式
type WSType struct {
	Encryption
	Conn *websocket.Conn
}

func (socket *WSType) Write(data []byte) (int, error) {
	socket.Encrypt(data)
	err := socket.Conn.WriteMessage(websocket.BinaryMessage, data)
	return len(data), err
}

func (socket *WSType) Read(data []byte) (int, error) {
	_, msg, err := socket.Conn.ReadMessage()
	if err != nil {
		return 0, err
	}
	socket.Decrypt(msg)
	copy(data, msg)
	return len(msg), nil
}

func (socket *WSType) Close() error {
	return socket.Conn.Close()
}

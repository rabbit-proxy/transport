package transport

import (
	"github.com/gorilla/websocket"
	"net"
	"time"
)

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

func (socket *TCPSocketType) LocalAddr() net.Addr {
	return socket.Conn.LocalAddr()
}

func (socket *TCPSocketType) RemoteAddr() net.Addr {
	return socket.Conn.RemoteAddr()
}

func (socket *TCPSocketType) SetDeadline(t time.Time) error {
	return socket.Conn.SetDeadline(t)
}

func (socket *TCPSocketType) SetReadDeadline(t time.Time) error {
	return socket.Conn.SetReadDeadline(t)
}

func (socket *TCPSocketType) SetWriteDeadline(t time.Time) error {
	return socket.Conn.SetWriteDeadline(t)
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

func (socket *WSType) LocalAddr() net.Addr {
	return socket.Conn.LocalAddr()
}

func (socket *WSType) RemoteAddr() net.Addr {
	return socket.Conn.RemoteAddr()
}

func (socket *WSType) SetDeadline(t time.Time) error {
	err := socket.Conn.SetReadDeadline(t)
	if err != nil {
		return err
	}
	return socket.Conn.SetWriteDeadline(t)
}

func (socket *WSType) SetReadDeadline(t time.Time) error {
	return socket.Conn.SetReadDeadline(t)
}

func (socket *WSType) SetWriteDeadline(t time.Time) error {
	return socket.Conn.SetWriteDeadline(t)
}
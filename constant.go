package transport

import (
	"context"
	"net"
	"time"
)

type SocketInterface interface {
	Read(ctx context.Context, data []byte) (int, error)
	Write(ctx context.Context, data []byte) (int, error)
	Close() error
}

type ConnSocketType struct {
	conn net.Conn
}

func (c *ConnSocketType) Read(ctx context.Context, data []byte) (int, error) {
	readCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		select {
		case <-ctx.Done():
			_ = c.conn.SetReadDeadline(time.Now())
		case <-readCtx.Done():
			return
		}
	}()

	return c.conn.Read(data)
}

func (c *ConnSocketType) Write(ctx context.Context, data []byte) (int, error) {
	writeCtx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	go func() {
		select {
		case <-ctx.Done():
			_ = c.conn.SetWriteDeadline(time.Now())
		case <-writeCtx.Done():
			return
		}
	}()
	return c.conn.Write(data)
}

func (c *ConnSocketType) Close() error {
	return c.conn.Close()
}

func NewConnSocket(conn net.Conn) *ConnSocketType {
	return &ConnSocketType{
		conn: conn,
	}
}

// CryptSocketType 加密连接
type CryptSocketType struct {
	socket SocketInterface
	crypt  Encryption
}

func (c *CryptSocketType) Read(ctx context.Context, data []byte) (int, error) {
	n, err := c.socket.Read(ctx, data)
	if err != nil {
		return n, err
	}
	c.crypt.Decrypt(data[:n])
	return n, nil
}

func (c *CryptSocketType) Write(ctx context.Context, data []byte) (int, error) {
	c.crypt.Encrypt(data)
	return c.socket.Write(ctx, data)
}

func (c *CryptSocketType) Close() error {
	return c.socket.Close()
}

func NewCryptSocket(socket SocketInterface, encryption Encryption) *CryptSocketType {
	return &CryptSocketType{
		socket:     socket,
		crypt: encryption,
	}
}

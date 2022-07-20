package transport

import (
	"io"
)

// CryptConnType 加密连接
type CryptConnType struct {
	conn  io.ReadWriteCloser
	crypt Encryption
}

func (c *CryptConnType) Read(data []byte) (int, error) {
	n, err := c.conn.Read(data)
	if err != nil {
		return n, err
	}
	c.crypt.Decrypt(data[:n])
	return n, nil
}

func (c *CryptConnType) Write(data []byte) (int, error) {
	c.crypt.Encrypt(data)
	return c.conn.Write(data)
}

func (c *CryptConnType) Close() error {
	return c.conn.Close()
}

func NewCryptSocket(conn io.ReadWriteCloser, encryption Encryption) *CryptConnType {
	return &CryptConnType{
		conn:  conn,
		crypt: encryption,
	}
}

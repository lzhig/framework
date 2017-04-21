package network

import "net"

type Connection struct {
	conn     net.Conn
	binddata interface{}
}

func (conn Connection) RemoteAddr() string {
	return conn.conn.RemoteAddr().String()
}

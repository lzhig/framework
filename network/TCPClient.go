package network

import (
	"fmt"
	"net"
	"time"
)

type DisconnectedCallbackT func(addr string, err error)
type MessageCallbackT func(packet *Packet)

type TCPClient struct {
	addr        string
	conn        Connection
	readBuffer  [1024 * 16]byte
	writeBuffer [1024 * 16]byte

	OnServerDisconnected DisconnectedCallbackT
	OnServerMessage      MessageCallbackT
}

func (c *TCPClient) Connect(addr string, timeout uint32, OnServerDisconnected DisconnectedCallbackT, OnServerMessage MessageCallbackT) (err error) {
	c.addr = addr
	conn, err := net.DialTimeout("tcp", addr, time.Millisecond*time.Duration(timeout))
	if err != nil {
		return err
	}

	c.conn = Connection{conn: conn}
	c.OnServerDisconnected = OnServerDisconnected
	c.OnServerMessage = OnServerMessage

	disconnectFunc := func(err error) {
		conn.Close()

		if c.OnServerDisconnected != nil {
			c.OnServerDisconnected(c.addr, err)
		}
	}

	// read goroutine
	go func() {
		var dataBegin int32
		var read int32
		p := Packet{}
		bufLen := int32(len(c.readBuffer))
		for {
			n, err := conn.Read(c.readBuffer[dataBegin:])
			if err != nil {
				disconnectFunc(err)
				return
			}

			if n > 0 {
				//fmt.Println(c.readBuffer[dataBegin : dataBegin+int32(n)])
				dataBegin += int32(n)
				for read < dataBegin {
					//fmt.Println("[", read, ":", dataBegin, "]")
					ok, headerLen, packetLen, err := packetHeader.ParsePacketHeader(c.readBuffer[read:dataBegin])
					//glog.Info(ok, headerLen, packetLen, err)
					if ok && dataBegin-read >= headerLen+packetLen {
						if c.OnServerMessage != nil {
							//glog.Info(packetLen)
							//glog.Info(b[read+headerLen : read+headerLen+packetLen])
							p.Attach(c.readBuffer[read+headerLen : read+headerLen+packetLen])
							//fmt.Println(p)
							c.OnServerMessage(&p)
						}
						read += headerLen + packetLen
					} else if err != nil {
						if e, ok := err.(*ErrorInvalidPacketHeader); e != nil && ok {
							disconnectFunc(e)
							return
						}
					} else if !ok || dataBegin-read < headerLen+packetLen {
						//glog.Info("dataBegin:", dataBegin, ", read: ", read)
						if read+headerLen > bufLen || dataBegin == bufLen {
							copy(c.readBuffer[:], c.readBuffer[read:dataBegin])
							dataBegin = dataBegin - read
							read = 0
						}
						break
					} else if headerLen+packetLen > bufLen {
						disconnectFunc(&ErrorPacketSizeTooLarge{ErrorNetwork{s: "Packet size is too large"}})
						return
					} else {
						disconnectFunc(&ErrorNetwork{s: "TCPServer: Logic Error, Assert!!!!!!"})
						return
					}
				}
			}
		}
	}()

	return nil
}

func (c *TCPClient) Disconnect() error {
	return c.conn.conn.Close()
}

func (c *TCPClient) Send(data []byte) {
	c.conn.conn.Write(data)
}

func (c *TCPClient) SendPacket(packet *Packet) (int, error) {
	buf := make([]byte, packet.GetPacketLen()+packetHeader.GetHeaderLen())
	packetHeader.BuildHeader(packet.GetPacketLen(), buf)
	copy(buf[packetHeader.GetHeaderLen():], packet.GetData())
	fmt.Println("send len:", len(buf))
	return c.conn.conn.Write(buf)
}

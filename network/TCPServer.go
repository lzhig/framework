package network

import (
	"net"
	"sync"

	"fmt"

	"github.com/golang/glog"
)

type clientConnections struct {
	connections map[net.Conn]*Connection
	mutex       sync.Mutex
}

func (ccs *clientConnections) init(n uint32) {
	ccs.connections = make(map[net.Conn]*Connection, n)
}

func (ccs *clientConnections) getConnectionsNumber() uint32 {
	return uint32(len(ccs.connections))
}

func (ccs *clientConnections) add(conn *Connection) {
	ccs.mutex.Lock()
	defer ccs.mutex.Unlock()

	ccs.connections[conn.conn] = conn
}

func (ccs *clientConnections) remove(conn *Connection) {
	ccs.mutex.Lock()
	defer ccs.mutex.Unlock()
	delete(ccs.connections, conn.conn)
}

type TCPServer struct {
	netListener       *net.TCPListener
	maxClients        uint32
	clientConnections clientConnections
	stopCmdChan       chan int32
	exitLoopChan      chan int32

	readBuffer [1024 * 16]byte

	onClientConnected    func(conn *Connection)
	onClientDisconnected func(conn *Connection, err error)
	onClientMessage      func(conn *Connection, packet *Packet)
}

func (s *TCPServer) Start(addr string, maxclients uint32,
	onClientConnected func(conn *Connection),
	onClientDisconnected func(conn *Connection, err error), onClientMessage func(conn *Connection, packet *Packet)) (err error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}
	s.netListener, err = net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return
	}

	s.maxClients = maxclients
	s.clientConnections.init(maxclients)
	s.onClientConnected = onClientConnected
	s.onClientDisconnected = onClientDisconnected
	s.onClientMessage = onClientMessage

	s.stopCmdChan = make(chan int32, 1)
	s.exitLoopChan = make(chan int32, 1)

	go s.loop()

	return nil
}

func (s *TCPServer) Stop() {
	s.stopCmdChan <- 0
	<-s.exitLoopChan
	s.netListener = nil
}

func (s *TCPServer) Disconnect(conn *Connection) error {
	glog.Info("Disconnect")
	err := conn.conn.Close()
	s.clientConnections.remove(conn)
	return err
}

func (s *TCPServer) Send(conn *Connection, data []byte) (n int, err error) {
	return conn.conn.Write(data)
}

func (s *TCPServer) SendPacket(conn *Connection, packet *Packet) (n int, err error) {
	buf := make([]byte, packet.GetPacketLen()+packetHeader.GetHeaderLen())
	packetHeader.BuildHeader(packet.GetPacketLen(), buf)
	copy(buf[packetHeader.GetHeaderLen():], packet.GetData())
	//glog.Info(buf)
	return conn.conn.Write(buf)
}

func (s *TCPServer) SetBindData(conn *Connection, data interface{}) {
	conn.binddata = data
}

func (s *TCPServer) GetBindData(conn *Connection) interface{} {
	return conn.binddata
}

func (s *TCPServer) loop() {
	defer s.netListener.Close()

	for {
		select {
		case <-s.stopCmdChan:
			s.exitLoopChan <- 0
			return
		default:
			//s.netListener.SetDeadline(time.Now().Add(time.Millisecond))
			conn, err := s.netListener.AcceptTCP()
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			} else if err != nil {
				glog.Fatal(err)
				return
			}
			if s.clientConnections.getConnectionsNumber() >= s.maxClients {
				conn.Close()
				continue
			}

			go s.connectionLoop(conn)
		}
	}
}

func (s *TCPServer) connectionLoop(conn *net.TCPConn) {
	c := &Connection{conn: conn}
	s.clientConnections.add(c)
	if s.onClientConnected != nil {
		s.onClientConnected(c)
	}

	bufLen := int32(len(s.readBuffer))

	conn.SetReadBuffer(int(bufLen))
	conn.SetWriteBuffer(int(bufLen))

	p := Packet{}

	var dataBegin int32
	var read int32

	disconnectFunc := func(err error) {
		if s.onClientDisconnected != nil {
			s.onClientDisconnected(c, err)
		}
		s.clientConnections.remove(c)
		c.conn.Close()
	}

	for {
		//time.Sleep(time.Nanosecond)
		n, err := conn.Read(s.readBuffer[dataBegin:])
		if err != nil {
			//glog.Info("dataBegin:", dataBegin, ", n:", n, ", err:", reflect.TypeOf(err))
			/*
				if e, ok := err.(*net.OpError); ok {
					glog.Info(reflect.TypeOf(e.Err))
					if e1, ok := e.Err.(*os.SyscallError); ok {
						glog.Info("type:", reflect.TypeOf(e1.Err))
						e2, _ := e1.Err.(syscall.Errno)
						glog.Info("type", e2.Timeout(), e2.Temporary(), e2.Error())
					}
					glog.Info("timeout:", e.Timeout(), e.Temporary())
					if e.Timeout() {
						continue
					}
				}*/

			disconnectFunc(err)
			return
		}
		//glog.Info("Read: ", n)

		if n > 0 {
			fmt.Println("Read:", n)
			//glog.Info(b[dataBegin : dataBegin+int32(n)])
			dataBegin += int32(n)
			for {
				//glog.Info("[", read, ":", dataBegin, "]")
				ok, headerLen, packetLen, err := packetHeader.ParsePacketHeader(s.readBuffer[read:dataBegin])
				//glog.Info(ok, headerLen, packetLen, err)
				if ok && dataBegin-read >= headerLen+packetLen {
					if s.onClientMessage != nil {
						//glog.Info(packetLen)
						//glog.Info(b[read+headerLen : read+headerLen+packetLen])
						p.Attach(s.readBuffer[read+headerLen : read+headerLen+packetLen])
						s.onClientMessage(c, &p)
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
						copy(s.readBuffer[:], s.readBuffer[read:dataBegin])
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
}

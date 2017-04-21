package main

import (
	"flag"
	"fmt"
	"globaltedinc/framework/glog"
	"globaltedinc/framework/network"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"runtime/pprof"
)

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	p := pprof.Lookup("goroutine")
	p.WriteTo(w, 1)
}

func onClientConnected(conn *network.Connection) {
	glog.Info("client " + conn.RemoteAddr() + " connected")
}

func onClientDisconnected(conn *network.Connection, err error) {
	glog.Info("client " + conn.RemoteAddr() + " disconnected")
	glog.Info(err)
}

func main() {
	flag.Parse()
	defer glog.Flush()

	/*go func() {
		log.Println(http.ListenAndServe("localhost:6061", nil))
	}()*/

	runtime.GOMAXPROCS(runtime.NumCPU())

	var s network.TCPServer

	err := s.Start("0.0.0.0:7890", 1000, onClientConnected, onClientDisconnected,
		func(conn *network.Connection, packet *network.Packet) {
			//glog.Info("client message receive")
			s.SendPacket(conn, packet)
		})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer s.Stop()

	select {}
}

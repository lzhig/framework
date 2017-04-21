package main

import (
	"flag"
	"fmt"
	"globaltedinc/framework/network"
	_ "net/http/pprof"
)

var packet = network.Packet{}

func OnServerMessage(packet network.Packet) {
	fmt.Println("server message.")
}

func connect(c *network.TCPClient) {
	for {

		fmt.Println("connecting...")
		err := c.Connect("10.63.7.20:7890", 2000,

			func(err error) {
				fmt.Println("server disconnected. error:", err)
				connect(c)
			},

			func(packet *network.Packet) {
				//fmt.Println("server message.")
				c.SendPacket(packet)
			})

		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println("connected.")

		n, err := c.SendPacket(&packet)
		fmt.Println(n, err)

		break
	}
}

func main() {
	flag.Parse()

	/*go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()*/

	fmt.Println("Starting...")

	var c network.TCPClient
	data := [4086]byte{}
	l := len(data)
	for i := 0; i < l/4; i++ {
		data[i*4] = byte((i & 0xff000000) >> 24)
		data[i*4+1] = byte((i & 0xff0000) >> 16)
		data[i*4+2] = byte((i & 0xff00) >> 8)
		data[i*4+3] = byte(i & 0xff)
	}

	packet.Attach(data[:])

	connect(&c)
	defer c.Disconnect()

	select {}
}

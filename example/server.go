package main

import (
	"github.com/andyzhou/cree"
	"github.com/andyzhou/cree/iface"
	"log"
)

/*
 * server
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

func OnConnAdd(conn iface.IConnect) {
	log.Println("add conn:", conn)
}

func OnConnLost(conn iface.IConnect) {
	log.Println("lost conn:", conn)
}

func main() {
	host := "127.0.0.1"
	port := 7800

	//init server
	server := cree.NewServer(host, port, "tcp4")

	//register hook for tcp connect start and stop
	server.SetOnConnStart(OnConnAdd)
	server.SetOnConnStop(OnConnLost)

	//setting for performance
	server.SetMaxConn(100)
	server.SetHandlerQueues(32)

	//register router for message
	server.AddRouter(1, &TestApi{})

	//start service
	server.Start()
}

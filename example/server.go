package main

import (
	"fmt"
	"github.com/andyzhou/cree"
	"github.com/andyzhou/cree/face"
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

	//init cb api
	testApi := NewTestApi()

	//init server
	server := cree.NewServer(host, port, "tcp")

	//register hook for tcp connect start and stop
	server.SetOnConnStart(OnConnAdd)
	server.SetOnConnStop(OnConnLost)

	//setting for performance
	server.SetMaxConn(100)
	server.SetHandlerQueues(32)

	//register router for message
	server.AddRouter(1, testApi)

	fmt.Printf("start server on %s:%d\n", host, port)
	//start service
	server.Start()
}

////////////
//api face
////////////

//api face
type TestApi struct {
	face.BaseRouter
}

//construct
func NewTestApi() *TestApi {
	this := &TestApi{}
	return this
}

func (*TestApi) Handle(req iface.IRequest) {
	log.Println(
		"TestApi::Handle, data:",
		string(req.GetMessage().GetData()),
	)
	message := req.GetMessage()
	req.GetConnect().SendMessage(
		message.GetId(),
		message.GetData(),
	)
}
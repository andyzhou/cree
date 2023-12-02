package main

import (
	"fmt"
	"github.com/andyzhou/cree"
	"github.com/andyzhou/cree/face"
	"github.com/andyzhou/cree/iface"
	"log"
	"os"
	"runtime/pprof"
	"time"
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
	//setup
	host := "127.0.0.1"
	port := 7800

	//cpu pprof
	f, _ := os.OpenFile("cpu.pprof", os.O_CREATE|os.O_RDWR, 0644)
	pprof.StartCPUProfile(f)

	sf := func() {
		pprof.StopCPUProfile()
		f.Close()
		fmt.Println("pprof cpu finished")
	}
	time.AfterFunc(time.Second * 15, sf)


	//set server conf
	conf := &cree.ServerConf{
		Host: host,
		Port: port,
		TcpVersion: "tcp",
		MaxConnects: 128,
	}

	//init server
	server := cree.NewServer(conf)

	//register hook for tcp connect start and stop
	server.SetOnConnStart(OnConnAdd)
	server.SetOnConnStop(OnConnLost)

	//setting for performance
	server.SetMaxConnects(100)
	//server.SetHandlerQueues(1)

	//init cb api
	testApi := NewTestApi()

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
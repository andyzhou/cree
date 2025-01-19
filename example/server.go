package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/pprof"
	"time"

	"github.com/andyzhou/cree"
	"github.com/andyzhou/cree/face"
	"github.com/andyzhou/cree/iface"
	_ "net/http/pprof"
)

/*
 * server
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

func OnConnAdd(conn iface.IConnect) {
	log.Println("add conn:", conn.GetConnId())
}

func OnConnLost(conn iface.IConnect) {
	log.Println("lost conn:", conn.GetConnId())
}

func OnReceiveMsg(conn iface.IConnect, req iface.IRequest) error {
	if conn == nil || req == nil {
		return errors.New("invalid parameter")
	}
	log.Printf("OnReceiveMsg, req:%v\n", req.GetMessage().GetId())
	return nil
}

//cpu pprof
func cpuPprof()  {
	//cpu pprof
	f, _ := os.OpenFile("cpu.pprof", os.O_CREATE|os.O_RDWR, 0644)
	pprof.StartCPUProfile(f)

	sf := func() {
		pprof.StopCPUProfile()
		f.Close()
		fmt.Println("pprof cpu finished")
	}
	time.AfterFunc(time.Second * 20, sf)
}

//run pprof
func runPProf()  {
	ip := "0.0.0.0:8080"
	if err := http.ListenAndServe(ip, nil); err != nil {
		fmt.Printf("start pprof failed on %s\n", ip)
	}
}

func main() {
	//setup
	host := "127.0.0.1"
	port := 7800

	//set server conf
	conf := &cree.ServerConf{
		Host: host,
		Port: port,
		TcpVersion: "tcp",
		BucketReadRate: 0.05,
		ErrMsgId: 100,
	}

	//start pprof
	//go runPProf()

	//init server
	server := cree.NewServer(conf)

	//register hook for new tcp connect start and stop
	server.SetConnected(OnConnAdd)
	server.SetDisconnected(OnConnLost)
	server.SetReadMessage(OnReceiveMsg)

	//setting for performance
	//server.SetMaxConnects(100)

	//init cb api
	testApi := NewTestApi()

	//register router for message
	server.AddRouter(1, testApi)
	server.AddRouter(2, testApi)
	server.AddRouter(3, testApi)

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
	//log.Println(
	//	"TestApi::Handle, data:",
	//	string(req.GetMessage().GetData()),
	//)
	message := req.GetMessage()
	req.GetConnect().SendMessage(
		message.GetId(),
		message.GetData(),
	)
}
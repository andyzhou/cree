package cree

import (
	"errors"
	"fmt"
	"log"
	"net"
	"runtime"
	"sync"
	"time"

	"github.com/andyzhou/cree/define"
	"github.com/andyzhou/cree/face"
	"github.com/andyzhou/cree/iface"
)

/*
 * face for server
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//server config
type ServerConf struct {
	Host         string
	Port         int
	TcpVersion   string //like tcp, tcp4, tcp6
	MaxConnects  int32
	MaxPackSize  int //pack data max size
	Buckets      int //bucket size for tcp connect
	BucketReadRate float64 //bucket data read rate
	LittleEndian bool
	GCRate		 int //xx seconds
}

 //face info
type Server struct {
	//basic
	conf         *ServerConf
	needQuit     bool
	littleEndian bool
	packet       iface.IPacket
	handler      iface.IHandler

	//hook
	onConnStart func(iface.IConnect)
	onConnStop  func(iface.IConnect)

	//others
	gcCloseChan chan bool
	gcTicker *time.Ticker
	wg sync.WaitGroup
	sync.RWMutex
}

 //construct
 //extraParas, first is tcp kind, second is max connects.
func NewServer(configs ...*ServerConf) *Server {
	var (
		conf *ServerConf
	)
	if configs != nil && len(configs) > 0 {
		conf = configs[0]
	}

	//check conf
	if conf == nil {
		conf = &ServerConf{
			Port: define.DefaultPort,
			TcpVersion: define.DefaultTcpVersion,
		}
	}
	if conf.TcpVersion == "" {
		conf.TcpVersion = define.DefaultTcpVersion
	}
	if conf.Port <= 0 {
		conf.Port = define.DefaultPort
	}
	if conf.GCRate <= 0 {
		conf.GCRate = define.DefaultGCRate
	}

	if conf.Buckets <= 0 {
		conf.Buckets = define.DefaultBuckets
	}
	if conf.BucketReadRate <= 0 {
		conf.BucketReadRate = define.DefaultBucketReadRate
	}

	//self init
	this := &Server{
		conf: conf,
		packet: face.NewPacket(),
		handler: face.NewHandler(),
		gcCloseChan: make(chan bool, 1),
	}

	//inter init
	this.interInit()
	return this
}

//start
func (s *Server) Start() {
	s.wg.Add(1)
	s.wg.Wait()
}

//stop
func (s *Server) Stop() {
	s.needQuit = true
	face.GetManager().Clear()
	s.gcCloseChan <- true
	s.wg.Done()
}

func (s *Server) StopSkipWg() {
	s.needQuit = true
	face.GetManager().Clear()
}

//add router for one message id
func (s *Server) AddRouter(messageId uint32, router iface.IRouter) {
	s.handler.AddRouter(messageId, router)
}

//register redirect router
//used for unsupported message id process
func (s *Server) RegisterRedirect(router iface.IRouter) {
	s.handler.RegisterRedirect(router)
}

//get conn manager
func (s *Server) GetManager() iface.IManager {
	return face.GetManager()
}

func (s *Server) GetPacket() iface.IPacket {
	return s.packet
}

//set max pack size
func (s *Server) SetMaxPackSize(size int) {
	s.packet.SetMaxPackSize(size)
}

func (s *Server) SetLittleEndian(littleEndian bool) {
	s.littleEndian = littleEndian
}

//set max connections
func (s *Server) SetMaxConnects(maxConnects int32) {
	if maxConnects <= 0 {
		return
	}
	s.conf.MaxConnects = maxConnects
}

//set hook
func (s *Server) SetOnConnStart(hook func(iface.IConnect)) {
	s.onConnStart = hook
}
func (s *Server) SetOnConnStop(hook func(iface.IConnect)) {
	s.onConnStop = hook
}

//call hook
func (s *Server) CallOnConnStart(conn iface.IConnect) {
	if s.onConnStart != nil {
		s.onConnStart(conn)
	}
}
func (s *Server) CallOnConnStop(conn iface.IConnect) {
	if s.onConnStop != nil {
		s.onConnStop(conn)
	}
}

////////////////
//private func
////////////////

//watch new tcp connect
func (s *Server) watchConn(listener *net.TCPListener) error {
	var (
		connId uint32
		m any = nil
	)

	//check
	if listener == nil {
		return errors.New("cree.server, listener not init")
	}

	//defer
	defer func() {
		if subErr := recover(); subErr != m {
			log.Println("cree.server, watch connect panic err:", subErr)
		}
	}()

	//loop
	for {
		if s.needQuit {
			break
		}

		//check max connects
		realConnects := face.GetManager().GetLen()
		if s.conf.MaxConnects > 0 &&
			realConnects >= s.conf.MaxConnects {
			log.Printf("cree.server, connect up to max count, config max conn:%v, real conn:%v\n",
				s.conf.MaxConnects, realConnects)
			continue
		}

		//get tcp connect
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Println("cree.server, accept connect failed, err:", err.Error())
			continue
		}

		//process new connect
		connId++
		connect := face.NewConnect(s, conn, connId, s.handler)

		//add connect into manager
		s.GetManager().Add(connect)

		//push into bucket
		face.GetBucket().AddConnect(connect)
	}
	return nil
}

//cb for new connect list consume
func (s *Server) cbForNewConnConsume(data interface{}) (interface{}, error) {
	//check
	if data == nil {
		return nil, errors.New("invalid parameter")
	}
	conn, ok := data.(*face.Connect)
	if !ok || conn == nil {
		return nil, errors.New("data should be `*Connect` type")
	}
	//start connect
	conn.Start()
	return nil, nil
}

//run gc ticker
func (s *Server) runGCTicker() {
	//defer
	defer func() {
		s.gcTicker.Stop()
	}()

	//loop
	for {
		select {
		case <- s.gcTicker.C:
			{
				runtime.GC()
			}
		case <- s.gcCloseChan:
			{
				return
			}
		}
	}
}

//inter init
func (s *Server) interInit() bool {
	//get tcp addr
	address := fmt.Sprintf("%s:%d", s.conf.Host, s.conf.Port)
	addr, err := net.ResolveTCPAddr(s.conf.TcpVersion, address)
	if err != nil {
		log.Printf("cree.server, resolve tcp addr failed, err:%v", err.Error())
		return false
	}

	//begin listen
	listener, subErr := net.ListenTCP(s.conf.TcpVersion, addr)
	if subErr != nil {
		log.Printf("cree.server, listen on %v failed, err:%v", address, subErr.Error())
		return false
	}

	//init manager instance
	face.GetManager()

	//setup bucket
	//face.GetBucket().SetCBForReadMessage()
	face.GetBucket().CreateSonWorkers(s.conf.Buckets, s.conf.BucketReadRate)

	//start gc ticker
	gcRate := time.Duration(s.conf.GCRate) * time.Second
	s.gcTicker = time.NewTicker(gcRate)
	go s.runGCTicker()

	//watch tcp connect
	go s.watchConn(listener)
	return true
}
package cree

import (
	"fmt"
	"github.com/andyzhou/cree/define"
	"github.com/andyzhou/cree/face"
	"github.com/andyzhou/cree/iface"
	"log"
	"net"
	"sync"
)

/*
 * face for server
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//server config
type ServerConf struct {
	Host string
	Port int
	TcpVersion string //like tcp, tcp4, tcp6
	MaxConnects int32
}

 //face info
 type Server struct {
 	//basic
 	conf *ServerConf
 	needQuit bool
 	littleEndian bool
 	packet iface.IPacket
 	handler iface.IHandler
 	manager iface.IManager
 	//hook
	onConnStart func(iface.IConnect)
 	onConnStop func(iface.IConnect)
 	wg sync.WaitGroup
 	sync.RWMutex
 }

 //construct
 //extraParas, first is tcp kind, second is max connects.
func NewServer(conf *ServerConf) *Server {
	//check conf
	if conf == nil {
		conf = &ServerConf{
			TcpVersion: define.DefaultTcpVersion,
			MaxConnects: define.DefaultMinConnects,
		}
	}

	//self init
	this := &Server{
		conf: conf,
		handler:face.NewHandler(),
		manager:face.NewManager(),
		packet: face.NewPacket(),
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
	s.wg.Done()
	s.manager.Clear()
}

func (s *Server) StopSkipWg() {
	s.needQuit = true
	s.manager.Clear()
}

//add router
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
	return s.manager
}

//set max connections
func (s *Server) SetMaxConnects(maxConnects int32) {
	if maxConnects <= 0 {
		return
	}
	s.conf.MaxConnects = maxConnects
}

//set handler max queues
func (s *Server) SetHandlerQueues(maxQueues int) {
	if maxQueues <= 0 {
		return
	}
	s.handler.SetQueueSize(maxQueues)
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

func (s *Server) SetLittleEndian(littleEndian bool) {
	s.littleEndian = littleEndian
}

func (s *Server) GetPacket() iface.IPacket {
	return s.packet
}

////////////////
//private func
////////////////

//watch tcp connect
func (s *Server) watchConn(listener *net.TCPListener) bool {
	var (
		connId uint32
		conn *net.TCPConn
		err error
		m any = nil
	)

	if listener == nil {
		return false
	}

	//defer
	defer func() {
		if subErr := recover(); subErr != m {
			log.Println("Server:watchConn panic err:", subErr)
		}
	}()

	//loop
	for {
		if s.needQuit {
			break
		}

		//get tcp connect
		conn, err = listener.AcceptTCP()
		if err != nil {
			log.Println("accept failed, err:", err.Error())
			continue
		}

		//check max connects
		if s.manager.GetLen() >= s.conf.MaxConnects {
			log.Println("connect up to max count")
			conn.Close()
			continue
		}

		//process new connect
		connId++
		connect := face.NewConnect(s, conn, connId, s.handler)

		//add connect into manager
		s.GetManager().Add(connect)
	}
	return true
}

//inter init
func (s *Server) interInit() bool {
	//get tcp addr
	address := fmt.Sprintf("%s:%d", s.conf.Host, s.conf.Port)
	addr, err := net.ResolveTCPAddr(s.conf.TcpVersion, address)
	if err != nil {
		log.Println("resolve tcp addr failed, err:", err.Error())
		return false
	}

	//begin listen
	listener, err := net.ListenTCP(s.conf.TcpVersion, addr)
	if err != nil {
		log.Println("listen on ", address, " failed, err:", err.Error())
		return false
	}

	//watch tcp connect
	go s.watchConn(listener)

	return true
}
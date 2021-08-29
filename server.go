package cree

import (
	"fmt"
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

 //inter macro define
 const (
 	DefaultMaxConnects = 1024
 	DefaultIpVersion = "tcp"
 )

 //face info
 type Server struct {
 	//basic
 	ipVersion string //tcp4 or others
 	ip string
 	port int
 	maxConnect int
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
 //ipVersion like "tcp,tcp4,tcp6"
func NewServer(ipVersion, ip string, port int) *Server {
	//check and set default value
	if ipVersion == "" {
		ipVersion = DefaultIpVersion
	}

	//self init
	this := &Server{
		ipVersion:ipVersion,
		ip:ip,
		port:port,
		maxConnect:DefaultMaxConnects,
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
func (s *Server) SetMaxConnects(maxConnects int) {
	if maxConnects <= 0 {
		return
	}
	s.maxConnect = maxConnects
}

//set handler max queues
func (s *Server) SetHandlerQueues(maxQueues int) {
	if maxQueues <= 0 {
		return
	}
	s.handler.SetQueueSize(maxQueues)
}

func (s *Server) SetMaxConn(connects int) {
	if connects <= 0 {
		return
	}
	s.maxConnect = connects
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
	)

	if listener == nil {
		return false
	}

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
		if s.manager.GetLen() >= s.maxConnect {
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
	address := fmt.Sprintf("%s:%d", s.ip, s.port)
	addr, err := net.ResolveTCPAddr(s.ipVersion, address)
	if err != nil {
		log.Println("resolve tcp addr failed, err:", err.Error())
		return false
	}

	//begin listen
	listener, err := net.ListenTCP(s.ipVersion, addr)
	if err != nil {
		log.Println("listen on ", address, " failed, err:", err.Error())
		return false
	}

	//watch tcp connect
	go s.watchConn(listener)

	return true
}
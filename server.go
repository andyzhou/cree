package cree

import (
	"github.com/andyzhou/cree/iface"
	"github.com/andyzhou/cree/face"
	"net"
	"fmt"
	"log"
	"sync"
)

/*
 * face for server
 */

 //inter macro define
 const (
 	DefaultMaxConnects = 1024
 )

 //face info
 type Server struct {
 	//basic
 	ipVersion string //tcp4 or others
 	ip string
 	port int
 	maxConnect int
 	handler iface.IHandler
 	manager iface.IManager
 	//hook
	onConnStart func(iface.IConnect)
 	onConnStop func(iface.IConnect)
 	wg sync.WaitGroup
 }

 //construct
func NewServer(ipVersion, ip string, port int) *Server {
	//self init
	this := &Server{
		ipVersion:ipVersion,
		ip:ip,
		port:port,
		maxConnect:DefaultMaxConnects,
		handler:face.NewHandler(),
		manager:face.NewManager(),
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
	s.wg.Done()
	s.manager.Clear()
}

//add router
func (s *Server) AddRouter(messageId uint32, router iface.IRouter) {
	s.handler.AddRouter(messageId, router)
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
		face.NewConnect(s, conn, connId, s.handler)

		//spawn new process for new connection
		//go newConnect.Start()
	}
	return true
}

//inter init
func (s *Server) interInit() {
	//get tcp addr
	address := fmt.Sprintf("%s:%d", s.ip, s.port)
	addr, err := net.ResolveTCPAddr(s.ipVersion, address)
	if err != nil {
		log.Println("resolve tcp addr failed, err:", err.Error())
		return
	}

	//begin listen
	listener, err := net.ListenTCP(s.ipVersion, addr)
	if err != nil {
		log.Println("listen on ", address, " failed, err:", err.Error())
		return
	}

	//watch tcp connect
	go s.watchConn(listener)
}
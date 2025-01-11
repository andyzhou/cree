package cree

import (
	"errors"
	"fmt"
	"log"
	"net"
	"runtime/debug"
	"sync"
	"sync/atomic"

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
	ErrMsgId	 uint32
	Buckets      int //bucket size for tcp connect
	BucketReadRate float64 //bucket data read rate
	LittleEndian bool
	GCRate		 int //xx seconds
}

 //face info
type Server struct {
	//basic
	conf         *ServerConf
	connId		 int64
	connects     int32
	needQuit     bool
	littleEndian bool
	packet       iface.IPacket
	handler      iface.IHandler
	bucketMap    map[int]iface.IBucket

	//hook
	cbOfConnected    func(iface.IConnect)
	cbOfGenConnId	 func()int64

	//others
	wg sync.WaitGroup
	sync.RWMutex
}

//global variable
var (
	_server *Server
	_once *sync.Once
)

//get single instance
func GetServer(configs ...*ServerConf) *Server {
	_once.Do(func() {
		_server = NewServer(configs...)
	})
	return _server
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

	//check and set default conf
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
		bucketMap: map[int]iface.IBucket{},
		packet: face.NewPacket(),
		handler: face.NewHandler(),
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
}

func (s *Server) StopSkipWg() {
	s.needQuit = true
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
//hook for read message for buckets
func (s *Server) SetReadMessage(hook func(iface.IConnect, iface.IRequest) error) {
	if hook == nil {
		return
	}
	for _, v := range s.bucketMap {
		v.SetCBForReadMessage(hook)
	}
}

//hook for disconnected for buckets
func (s *Server) SetDisconnected(hook func(iface.IConnect)) {
	if hook == nil {
		return
	}
	//apply to sub buckets
	for _, v := range s.bucketMap {
		v.SetCBForDisconnected(hook)
	}
}

//hook for new connected for server
func (s *Server) SetConnected(hook func(iface.IConnect)) {
	s.cbOfConnected = hook
}

//hook for gen new connect id for server
func (s *Server) SetGenConnId(hook func()int64) {
	s.cbOfGenConnId = hook
}

////////////////
//private func
////////////////

//watch new tcp connect
func (s *Server) watchConn(listener *net.TCPListener) error {
	var (
		connId int64
		m any = nil
	)

	//check
	if listener == nil {
		return errors.New("cree.server, listener not init")
	}

	//defer
	defer func() {
		if subErr := recover(); subErr != m {
			log.Printf("cree.server, watch connect panic err:%v, trace:%v\n",
						subErr, string(debug.Stack()))
		}
	}()

	//loop
	for {
		//check max connects
		if s.conf.MaxConnects > 0 &&
			s.connects >= s.conf.MaxConnects {
			log.Printf("cree.server, connect up to max count, config max conn:%v, real conn:%v\n",
				s.conf.MaxConnects, s.connects)
			continue
		}

		//get tcp connect
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Println("cree.server, accept connect failed, err:", err.Error())
			continue
		}

		//process new connect
		//gen new connect id
		if s.cbOfGenConnId != nil {
			connId = s.cbOfGenConnId()
		}else{
			connId = atomic.AddInt64(&s.connId, 1)
		}
		if connId <= 0 {
			log.Println("can't gen new connect id")
			continue
		}

		//init new connect obj
		connect := face.NewConnect(s, conn, connId, s.handler)

		//call cb for new connected
		if s.cbOfConnected != nil {
			s.cbOfConnected(connect)
		}

		//push into target bucket
		bucket := s.getBucket(connId)
		bucket.AddConnect(connect)
	}
	return nil
}

//get bucket by connect id
func (s *Server) getBucket(connId int64) iface.IBucket {
	//check
	if connId <= 0 {
		return nil
	}

	//get target bucket id
	bucketId := int(connId % int64(s.conf.Buckets))

	//get target bucket
	s.Lock()
	defer s.Unlock()
	v, ok := s.bucketMap[bucketId]
	if ok && v != nil {
		return v
	}
	return nil
}

//inter init
func (s *Server) interInit() bool {
	//get tcp addr
	address := fmt.Sprintf("%s:%d", s.conf.Host, s.conf.Port)
	addr, err := net.ResolveTCPAddr(s.conf.TcpVersion, address)
	if err != nil {
		log.Printf("cree.server, resolve tcp addr failed, err:%v", err.Error())
		panic(any(err))
		return false
	}

	//begin listen
	listener, subErr := net.ListenTCP(s.conf.TcpVersion, addr)
	if subErr != nil {
		log.Printf("cree.server, listen on %v failed, err:%v", address, subErr.Error())
		panic(any(subErr))
		return false
	}

	//init inter buckets
	for i := 0; i < s.conf.Buckets; i++ {
		bucket := face.NewBucket(i, s.conf.ErrMsgId)
		s.bucketMap[i] = bucket
	}

	//watch tcp connect
	go s.watchConn(listener)
	return true
}
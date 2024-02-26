package face

import (
	"errors"
	"log"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/andyzhou/cree/define"
	"github.com/andyzhou/cree/iface"
)

/*
 * face for connect manager
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//global variables
var (
	_manager *Manager
	_managerOnce sync.Once
)

//inter type
type (
	msgCastReq struct {
		connId uint32
		data []byte
		forAll bool
	}
)

 //face info
 type Manager struct {
	connUnActiveSeconds int //max un-active seconds or force closed
	connectMap map[uint32]iface.IConnect
 	connects int32 //atomic value
	unActiveSeconds int
	tickChan chan struct{}
	closeChan chan bool
	sync.RWMutex
 }

 //get single instance
func GetManager() *Manager {
	_managerOnce.Do(func() {
		_manager = NewManager()
	})
	return _manager
}
 
 //construct
func NewManager() *Manager {
	//self init
	this := &Manager{
		unActiveSeconds: define.DefaultUnActiveSeconds,
		connectMap: map[uint32]iface.IConnect{},
		tickChan: make(chan struct{}, 1),
		closeChan: make(chan bool, 1),
	}

	//spawn son process
	go this.checkUnActiveConnProcess()
	return this
}

//quit
func (m *Manager) Quit() {
	if m.closeChan != nil {
		m.closeChan <- true
	}
}

 //set un-active seconds
func (m *Manager) SetUnActiveSeconds(val int) {
	if val <= 0 {
		return
	}
	m.unActiveSeconds = val
}

//clear all
func (m *Manager) Clear() {
	//basic check
	if m.connectMap == nil {
		return
	}
	//clear all
	m.Lock()
	defer m.Unlock()
	for k, v := range m.connectMap {
		if v != nil {
			v.Stop()
		}
		delete(m.connectMap, k)
	}
	m.connectMap = map[uint32]iface.IConnect{}
	atomic.StoreInt32(&m.connects, 0)

	//clean bucket
	GetBucket().Quit()
}

//get map length
func (m *Manager) GetLen() int32 {
	return m.connects
}

//get connect by id
func (m *Manager) Get(connId uint32) (iface.IConnect, error) {
	m.Lock()
	defer m.Unlock()
	v, ok := m.connectMap[connId]
	if !ok || v == nil {
		return nil, errors.New("connect not found")
	}
	return v, nil
}

//remove connect
func (m *Manager) Remove(conn iface.IConnect) error {
	//remove from map with locker
	if m.connectMap == nil {
		return errors.New("no any connections")
	}

	//check
	hasExists := m.connIsExists(conn.GetConnId())
	if hasExists {
		//found and delete it
		m.Lock()
		defer m.Unlock()
		delete(m.connectMap, conn.GetConnId())
		atomic.AddInt32(&m.connects, -1)

		//border value check
		if m.connects <= 0 {
			m.connectMap = map[uint32]iface.IConnect{}
			atomic.StoreInt32(&m.connects, 0)
			runtime.GC()
		}
	}

	//remove from bucket
	err := GetBucket().RemoveConnect(int64(conn.GetConnId()))
	return err
}

//add connect
func (m *Manager) Add(conn iface.IConnect) error {
	if conn == nil {
		return errors.New("invalid connect")
	}
	hasExists := m.connIsExists(conn.GetConnId())
	if hasExists {
		return errors.New("connect has exists")
	}
	//add into map with locker
	m.Lock()
	defer m.Unlock()
	m.connectMap[conn.GetConnId()] = conn
	atomic.AddInt32(&m.connects, 1)
	return nil
}

////////////////
//private func
////////////////

//check connect
func (m *Manager) connIsExists(connId uint32) bool {
	if connId <= 0 {
		return false
	}
	m.Lock()
	defer m.Unlock()
	_, ok := m.connectMap[connId]
	if ok {
		return true
	}
	return false
}

//check and kick un-active conn
func (m *Manager) checkUnActiveConn() {
	if m.connects <= 0 {
		return
	}

	//loop check
	needReset := false
	now := time.Now().Unix()
	m.Lock()
	defer m.Unlock()
	for k, v := range m.connectMap {
		lastActive := v.GetActiveTime()
		diff := int(now - lastActive)
		if diff >= m.unActiveSeconds {
			//found and remove it
			v.Stop()
			delete(m.connectMap, k)
			atomic.AddInt32(&m.connects, -1)
			needReset = true

			//remove from bucket
			GetBucket().RemoveConnect(int64(k))
		}
	}

	//border value check
	if m.connects <= 0 {
		atomic.StoreInt32(&m.connects, 0)
		if needReset {
			m.connectMap = map[uint32]iface.IConnect{}
			runtime.GC()
		}
	}
}

 //check un-active connect process
func (m *Manager) checkUnActiveConnProcess() {
	var (
		p any = nil
	)
	defer func() {
		if err := recover(); err != p {
			log.Printf("manager.panic, err:%v\n", err)
		}
	}()

	//start first ticker
	m.sendTicker(true)

	//loop
	for {
		select {
		case <- m.tickChan:
			{
				//check un-active conn
				m.checkUnActiveConn()

				//send next ticker
				m.sendTicker()
			}
		case <- m.closeChan:
			return
		}
	}
}

//send ticker
func (m *Manager) sendTicker(forces ...bool) {
	var (
		force bool
	)
	//check
	if forces != nil && len(forces) > 0 {
		force = forces[0]
	}
	if m.tickChan == nil {
		return
	}

	//send ticker
	ticker := func() {
		m.tickChan <- struct{}{}
	}
	if force {
		ticker()
	}else{
		time.AfterFunc(time.Second * define.DefaultManagerTicker, ticker)
	}
}

//cb for worker cast opt
func (m *Manager) cbForWorkerCastOpt(
		data interface{},
	) (interface{}, error) {
	//check
	if data == nil {
		return nil, errors.New("invalid parameter")
	}
	req, ok := data.(msgCastReq)
	if !ok || &req == nil {
		return nil, errors.New("data should be `msgCastReq` type")
	}

	//process request data
	connId := req.connId
	dataBytes := req.data

	//get connect
	conn, err := m.Get(connId)
	if err != nil || conn == nil {
		return nil, errors.New("can't get conn by id")
	}

	//send to connect client
	err = conn.SendMessage(connId, dataBytes)
	return nil, err
}
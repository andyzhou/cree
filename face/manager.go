package face

import (
	"errors"
	"github.com/andyzhou/cree/define"
	"github.com/andyzhou/cree/iface"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

/*
 * face for connect manager
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

 //face info
 type Manager struct {
	connUnActiveSeconds int //max un-active seconds or force closed
 	connectMap *sync.Map //connectId -> IConnect
 	connects int32 //atomic value
	unActiveSeconds int
	tickChan chan struct{}
	closeChan chan bool
 }
 
 //construct
func NewManager() *Manager {
	//self init
	this := &Manager{
		unActiveSeconds: define.DefaultUnActiveSeconds,
		connectMap:new(sync.Map),
		tickChan: make(chan struct{}, 1),
		closeChan: make(chan bool, 1),
	}
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

//get connect by id
func (m *Manager) Get(connId uint32) (iface.IConnect, error) {
	v, ok := m.connectMap.Load(connId)
	if !ok {
		return nil, errors.New("connect not found")
	}
	conn, ok := v.(iface.IConnect)
	if !ok {
		return nil, errors.New("invalid connect")
	}
	return conn, nil
}

//get map length
func (m *Manager) GetLen() int32 {
	return m.connects
}

//clear all
func (m *Manager) Clear() {
	//basic check
	if m.connectMap == nil {
		return
	}
	//clear all
	subFunc := func(key, val interface{}) bool {
		conn, ok := val.(iface.IConnect)
		if ok && conn != nil {
			conn.Stop()
		}
		return true
	}
	m.connectMap.Range(subFunc)
	m.connectMap = &sync.Map{}
	atomic.StoreInt32(&m.connects, 0)
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
		m.connectMap.Delete(conn.GetConnId())
		atomic.AddInt32(&m.connects, -1)
	}
	return nil
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
	m.connectMap.Store(conn.GetConnId(), conn)
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
	_, ok := m.connectMap.Load(connId)
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
	now := time.Now().Unix()
	sf := func(k, v interface{}) bool {
		//check
		connId, ok := k.(uint32)
		conn, okTwo := v.(iface.IConnect)
		if ok && okTwo && conn != nil {
			lastActive := conn.GetActiveTime()
			diff := int(now - lastActive)
			if diff >= m.unActiveSeconds {
				//found and remove it
				conn.Stop()
				m.connectMap.Delete(connId)
				atomic.AddInt32(&m.connects, -1)
			}
		}
		return true
	}
	m.connectMap.Range(sf)
	if m.connects < 0 {
		atomic.StoreInt32(&m.connects, 0)
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
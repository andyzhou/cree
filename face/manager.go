package face

import (
	"errors"
	"github.com/andyzhou/cree/iface"
	"sync"
)

/*
 * face for connect manager
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

 //face info
 type Manager struct {
 	connectMap *sync.Map //connectId -> IConnect
 	connects int
 	sync.RWMutex
 }
 
 //construct
func NewManager() *Manager {
	//self init
	this := &Manager{
		connectMap:new(sync.Map),
	}
	return this
}

//get connect by id
func (m *Manager) Get(connId uint32) (iface.IConnect, error) {
	//conn, ok := m.connectMap[connId]
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
func (m *Manager) GetLen() int {
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
		if !ok {
			return false
		}
		conn.Stop()
		return true
	}
	m.connectMap.Range(subFunc)
}

//remove
func (m *Manager) Remove(conn iface.IConnect) {
	//remove from map with locker
	if m.connectMap == nil {
		return
	}
	m.connectMap.Delete(conn.GetConnId())
}

//add connect
func (m *Manager) Add(conn iface.IConnect)  {
	if conn == nil {
		return
	}
	hasExists := m.connIsExists(conn.GetConnId())
	if hasExists {
		return
	}
	//add into map with locker
	m.Lock()
	defer m.Unlock()
	m.connectMap.Store(conn.GetConnId(), conn)
	m.connects++
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
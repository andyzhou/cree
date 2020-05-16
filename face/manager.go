package face

import (
	"github.com/andyzhou/cree/iface"
	"sync"
	"errors"
)

/*
 * face for connect manager
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

 //face info
 type Manager struct {
 	connectMap map[uint32]iface.IConnect `connectId -> IConnect`
 	sync.RWMutex
 }
 
 //construct
func NewManager() *Manager {
	//self init
	this := &Manager{
		connectMap:make(map[uint32]iface.IConnect),
	}
	return this
}

//get connect by id
func (m *Manager) Get(connId uint32) (iface.IConnect, error) {
	conn, ok := m.connectMap[connId]
	if ok {
		return conn, nil
	}

	return nil, errors.New("connect not found")
}

//get map length
func (m *Manager) GetLen() int {
	return len(m.connectMap)
}

//clear all
func (m *Manager) Clear() {
	//basic check
	if m.connectMap == nil || len(m.connectMap) <= 0 {
		return
	}

	//clear all
	for _, conn := range m.connectMap {
		conn.Stop()
	}
}

//remove
func (m *Manager) Remove(conn iface.IConnect) {
	//remove from map with locker
	m.Lock()
	defer m.Unlock()
	delete(m.connectMap, conn.GetConnId())
}

//add connect
func (m *Manager) Add(conn iface.IConnect)  {
	//add into map with locker
	m.Lock()
	defer m.Unlock()
	m.connectMap[conn.GetConnId()] = conn
}

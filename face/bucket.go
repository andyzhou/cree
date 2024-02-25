package face

import (
	"errors"
	"github.com/andyzhou/cree/iface"
	"sync"
	"time"
)

/*
 * connect bucket face
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 * - batch buckets contain all tcp connects
 * - loop read all sub connects
 */

//inter type
type (
	SonBucket struct {
		connectMap map[uint32]iface.IConnect
		connects int32 //atomic value
		readTickChan chan struct{}
		readCloseChan chan struct{}
		sync.RWMutex
	}
)

//face info
type Bucket struct {
	bucketMap map[int32]*SonBucket //groupId -> *SonBucket
	sync.RWMutex
}

//construct
func NewBucket() *Bucket {
	this := &Bucket{
		bucketMap: map[int32]*SonBucket{},
	}
	return this
}

/////////////////////
//api for son bucket
/////////////////////

//construct
func NewSonBucket(groupId int32) *SonBucket {
	this := &SonBucket{
		connectMap: map[uint32]iface.IConnect{},
		readTickChan: make(chan struct{}, 1),
		readCloseChan: make(chan struct{}, 1),
	}
	go this.runReadTicker()
	return this
}

//quit
func (f *SonBucket) Quit() {
	close(f.readCloseChan)
	time.Sleep(time.Second/5)
	f.Lock()
	defer f.Unlock()
	for k, v := range f.connectMap {
		v.Stop()
		delete(f.connectMap, k)
	}
}

//get connect
func (f *SonBucket) GetConn(connId uint32) (iface.IConnect, error) {
	//check
	if connId <= 0 {
		return nil, errors.New("invalid parameter")
	}
	//get with locker
	f.Lock()
	defer f.Unlock()
	v, ok := f.connectMap[connId]
	if ok && v != nil {
		return v, nil
	}
	return nil, errors.New("no connect by id")
}

//remove connect
func (f *SonBucket) RemoveConn(connId uint32) error {
	//check
	if connId <= 0 {
		return errors.New("invalid parameter")
	}
	//remove with locker
	f.Lock()
	defer f.Unlock()
	delete(f.connectMap, connId)
	return nil
}

//add connect
func (f *SonBucket) AddConn(conn iface.IConnect) error {
	//check
	if conn == nil {
		return errors.New("invalid parameter")
	}
	//add with locker
	f.Lock()
	defer f.Unlock()
	f.connectMap[conn.GetConnId()] = conn
	return nil
}

////////////////
//private func
////////////////

//run read ticker
func (f *SonBucket) runReadTicker() {

}
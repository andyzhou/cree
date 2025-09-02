package face

import (
	"errors"
	"github.com/andyzhou/cree/define"
	"github.com/andyzhou/cree/iface"
	"github.com/andyzhou/tinylib/queue"
	"io"
	"log"
	"math/rand"
	"runtime"
	"sync"
)

/*
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 * dynamic connect group
 * - dynamic create temp groups
 * - inter connections are all references
 */

//face info
type Group struct {
	groupId       int64
	errMsgId      uint32
	connMap       map[int64]iface.IConnect //reference conn map
	readMsgTicker *queue.Ticker            //ticker for read connect msg
	readMsgRate   float64

	//cb func
	cbForReadMessage  func(int64, iface.IConnect, iface.IRequest) error
	cbForDisconnected func(iface.IConnect)
	sync.RWMutex
}

//construct
func NewGroup(groupId int64, readMsgRates ...float64) *Group {
	var (
		readMsgRate float64
	)
	if readMsgRates != nil && len(readMsgRates) > 0 {
		readMsgRate = readMsgRates[0]
	}

	this := &Group{
		groupId: groupId,
		readMsgRate: readMsgRate,
		connMap: map[int64]iface.IConnect{},
	}
	this.interInit()
	return this
}

//clear
func (f *Group) Clear() {
	//stop read ticker
	if f.readMsgTicker != nil {
		f.readMsgTicker.Quit()
		f.readMsgTicker = nil
	}

	//inter data clean up
	f.Lock()
	defer f.Unlock()
	for k, conn := range f.connMap {
		conn.SetGroupId(0)
		delete(f.connMap, k)
	}
	f.connMap = nil
	runtime.GC()
}

//send message to all
func (f *Group) SendMessage(msgId uint32, msg []byte) error {
	var (
		err error
	)
	//check
	if msgId <= 0 || msg == nil || len(msg) <= 0 {
		return errors.New("invalid parameter")
	}

	//cast to all with locker
	f.Lock()
	defer f.Unlock()
	for _, conn := range f.connMap {
		err = conn.SendMessage(msgId, msg)
	}
	return err
}

//quit group
func (f *Group) Quit(connections ...iface.IConnect) error {
	//check
	if connections == nil || len(connections) <= 0 {
		return errors.New("invalid parameter")
	}

	//remove from map with locker
	f.Lock()
	defer f.Unlock()
	for _, conn := range connections {
		conn.SetGroupId(0)
		delete(f.connMap, conn.GetConnId())
	}

	//rebuild rate check
	rebuildRate := rand.Intn(define.FullPercent)
	needRebuild := false
	if rebuildRate > 0 && rebuildRate <= define.DefaultRebuildRate {
		needRebuild = true
	}

	//check and gc opt
	if needRebuild || len(f.connMap) <= 0 {
		newConnMap := map[int64]iface.IConnect{}
		for k, v := range f.connMap {
			newConnMap[k] = v
		}
		f.connMap = newConnMap
	}
	return nil
}

//join group
func (f *Group) Join(conn iface.IConnect) error {
	//check
	if conn == nil || conn.GetConnId() <= 0 {
		return errors.New("invalid parameter")
	}

	//sync into map with locker
	f.Lock()
	defer f.Unlock()
	conn.SetGroupId(f.groupId)
	f.connMap[conn.GetConnId()] = conn
	return nil
}

//set error message id
func (f *Group) SetErrMsgId(msgId uint32) {
	f.errMsgId = msgId
}

//set cb for read message
func (f *Group) SetCBForReadMessage(cb func(int64, iface.IConnect, iface.IRequest) error) {
	f.cbForReadMessage = cb
}

//set cb for disconnect
func (f *Group) SetCBForDisconnect(cb func(iface.IConnect))  {
	f.cbForDisconnected = cb
}

///////////////
//private func
///////////////

//close and remove connect
func (f *Group) closeConn(conn iface.IConnect) error {
	//check
	if conn == nil {
		return errors.New("invalid parameter")
	}

	//get key data
	connId := conn.GetConnId()

	//close connect and remove it
	conn.Quit()

	//check and call closed cb
	if f.cbForDisconnected != nil {
		f.cbForDisconnected(conn)
	}

	//rebuild rate check
	rebuildRate := rand.Intn(define.FullPercent)
	needRebuild := false
	if rebuildRate > 0 && rebuildRate <= define.DefaultRebuildRate {
		needRebuild = true
	}

	f.Lock()
	defer f.Unlock()
	delete(f.connMap, connId)
	if needRebuild || len(f.connMap) <= 0 {
		newConnMap := map[int64]iface.IConnect{}
		for k, v := range f.connMap {
			newConnMap[k] = v
		}
		f.connMap = newConnMap
	}
	return nil
}

//cb for read connect data
func (f *Group) cbForReadConnData() error {
	var (
		req iface.IRequest
		err error
	)
	//check
	if f.connMap == nil {
		return errors.New("no any active connections")
	}

	//loop read connect data with locker
	f.Lock()
	defer f.Unlock()
	for connId, conn := range f.connMap {
		//check connect
		if connId <= 0 || conn == nil {
			continue
		}

		//read message
		req, err = conn.ReadMessage()
		if err != nil {
			if err == io.EOF {
				//io read failed
				//close connect and remove it
				f.closeConn(conn)
			}else{
				//general error
				log.Printf("group %v read conn %v data failed, err:%v\n",
					f.groupId, conn.GetConnId(), err.Error())

				//send error to client connect
				conn.SendMessage(f.errMsgId, []byte(err.Error()))
			}
			continue
		}

		//check and call read message cb
		if f.cbForReadMessage != nil {
			f.cbForReadMessage(f.groupId, conn, req)
		}
	}
	return nil
}

//init read message ticker
func (f *Group) initReadMsgTicker() {
	//init read msg ticker
	f.readMsgTicker = queue.NewTicker(f.readMsgRate)
	f.readMsgTicker.SetCheckerCallback(f.cbForReadConnData)
}

//inter init
func (f *Group) interInit() {
	//check and set read message rate
	if f.readMsgRate <= 0 {
		f.readMsgRate = define.DefaultBucketReadRate
	}
	f.initReadMsgTicker()
}
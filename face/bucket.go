package face

import (
	"errors"
	"log"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/andyzhou/cree/define"
	"github.com/andyzhou/cree/iface"
	"github.com/andyzhou/tinylib/queue"
)

/*
 * connect bucket face
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 * - bucket id hashed by connect id
 * - batch buckets contain all tcp connects
 * - one bucket contain batch connects
 */

//face info
type Bucket struct {
	//inter obj
	bucketId      int
	readMsgTicker *queue.Ticker //ticker for read connect msg
	sendMsgQueue  *queue.List   //inter queue for send message

	//run env data
	connMap   map[int64]iface.IConnect //connId -> IConnect
	connCount int64

	//cb func
	cbForReadMessage func(iface.IConnect, iface.IRequest) error
	cbForDisconnected func(iface.IConnect)
	sync.RWMutex
}

//construct
func NewBucket(id int) *Bucket {
	this := &Bucket{
		bucketId: id,
		connMap: map[int64]iface.IConnect{},
	}
	this.interInit()
	return this
}

//quit
func (f *Bucket) Quit() {
	//close inter ticker
	if f.readMsgTicker != nil {
		f.readMsgTicker.Quit()
	}
	f.freeRunMemory()
}

//send sync message
func (f *Bucket) SendMessage(req *define.SendMsgReq) error {
	//check
	if req == nil || req.MsgId < 0 || req.Data == nil {
		return errors.New("invalid parameter")
	}
	if f.sendMsgQueue == nil {
		return errors.New("inter send message queue is nil")
	}

	//save into running queue
	err := f.sendMsgQueue.Push(req)
	return err
}

//////////////////
//api for cb func
//////////////////

//set cb for read message
func (f *Bucket) SetCBForReadMessage(cb func(iface.IConnect, iface.IRequest) error)  {
	if cb == nil {
		return
	}
	f.cbForReadMessage = cb
}

//set cb for conn disconnected
func (f *Bucket) SetCBForDisconnected(cb func(connect iface.IConnect)) {
	if cb == nil {
		return
	}
	f.cbForDisconnected = cb
}

//////////////////
//api for connect
//////////////////

//get connect by id
func (f *Bucket) GetConnect(connId int64) (iface.IConnect, error) {
	//check
	if connId <= 0 {
		return nil, errors.New("invalid parameter")
	}

	//get target with locker
	f.Lock()
	defer f.Unlock()
	conn, ok := f.connMap[connId]
	if ok && conn != nil {
		return conn, nil
	}
	return nil, nil
}

//remove connect
func (f *Bucket) RemoveConnect(connId int64) error {
	//check
	if connId <= 0 {
		return errors.New("invalid parameter")
	}

	//get target conn
	conn, _ := f.GetConnect(connId)
	if conn == nil {
		return errors.New("can't get conn by id")
	}

	//close and remove
	err := f.closeConn(conn)
	return err
}

//add new connect
func (f *Bucket) AddConnect(conn iface.IConnect) error {
	//check
	if conn == nil {
		return errors.New("invalid parameter")
	}

	//check old connect
	connId := int64(conn.GetConnId())
	oldConn, _ := f.GetConnect(connId)
	if oldConn != nil {
		return errors.New("conn already exists")
	}

	//sync into run env with locker
	f.Lock()
	defer f.Unlock()
	f.connMap[connId] = conn
	atomic.AddInt64(&f.connCount, 1)

	return nil
}

///////////////
//private func
///////////////

////cb for read connect data
//func (f *Bucket) cbForReadConnDataOld(
//	workerId int32,
//	connMaps ...interface{}) error {
//	var (
//		req iface.IRequest
//		err error
//	)
//	//check
//	if workerId <= 0 || connMaps == nil ||
//		len(connMaps) <= 0 {
//		return errors.New("invalid parameter")
//	}
//	mapVal := connMaps[0]
//	if mapVal == nil {
//		return errors.New("invalid conn map data")
//	}
//	connMap, ok := mapVal.(map[int64]interface{})
//	if !ok || connMap == nil || len(connMap) <= 0 {
//		return errors.New("no any conn map data")
//	}
//
//	//loop read connect data
//	for connId, conn := range connMap {
//		//check connect
//		if connId <= 0 || conn == nil {
//			continue
//		}
//		connObj, subOk := conn.(iface.IConnect)
//		if !subOk || connObj == nil {
//			continue
//		}
//
//		//read message
//		req, err = connObj.ReadMessage()
//		if err != nil {
//			//log.Printf("bucket.cbForReadConnData, read message failed, connId:%v, err:%v\n",
//			//	connId, err.Error())
//			connObj.Stop()
//			continue
//		}
//
//		//check and call rad message cb
//		if f.cbForReadMessage != nil {
//			f.cbForReadMessage(connId, req)
//		}
//	}
//	return err
//}

//cb for read connect data
func (f *Bucket) cbForReadConnData() error {
	var (
		req iface.IRequest
		err error
	)
	//check
	if f.connCount <= 0 || f.connMap == nil {
		return errors.New("no any active connections")
	}

	//loop read connect data
	for connId, conn := range f.connMap {
		//check connect
		if connId <= 0 || conn == nil {
			continue
		}

		//read message
		req, err = conn.ReadMessage()
		if err != nil {
			//close connect and remove it
			f.closeConn(conn)
			continue
		}

		//if bytes.Compare(f.router.GetHeartByte(), message) == 0 {
		//	//it's heart beat data
		//	connObj.HeartBeat()
		//	continue
		//}

		//check and call read message cb
		if f.cbForReadMessage != nil {
			f.cbForReadMessage(conn, req)
		}
	}

	return nil
}

//cb for send msg consumer
func (f *Bucket) cbForConsumerSendData(data interface{}) error {
	var (
		checkPass bool
		err error
	)
	//check
	if data == nil {
		return errors.New("invalid parameter")
	}

	//verify request data
	req, ok := data.(*define.SendMsgReq)
	if !ok || req == nil ||
		req.MsgId < 0 ||
		req.Data == nil {
		return errors.New("invalid parameter")
	}

	//loop send
	for _, v := range f.connMap {
		//check
		if v == nil {
			continue
		}
		//check send condition
		checkPass = f.checkSendCondition(req, v)
		if !checkPass {
			continue
		}
		//send message to target connect
		err = v.SendMessage(req.MsgId, req.Data)
		if err != nil {
			log.Printf("bucket.cbForConsumerSendData failed, err:%v\n", err.Error())
		}
	}
	return err
}

//check send condition
//if check pass, return true or false
func (f *Bucket) checkSendCondition(
	req *define.SendMsgReq,
	conn iface.IConnect) bool {
	//check
	if req == nil || conn == nil {
		return false
	}
	if req.ConnIds == nil &&
		req.Tags == nil {
		//not need condition check
		return true
	}

	//check by conn ids
	if len(req.ConnIds) > 0 {
		connOwnerId := int64(conn.GetConnId())
		if connOwnerId <= 0 {
			return false
		}
		for _, connId := range req.ConnIds {
			if connId == connOwnerId {
				return true
			}
		}
		return false
	}

	//check by tags
	if len(req.Tags) > 0 {
		connTags := conn.GetTags()
		if connTags == nil || len(connTags) <= 0 {
			return false
		}
		for _, tag := range req.Tags {
			if v, ok := connTags[tag]; ok && v {
				return true
			}
		}
		return false
	}
	return true
}

//close and remove connect
func (f *Bucket) closeConn(conn iface.IConnect) error {
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

	//remove from run env
	f.Lock()
	defer f.Unlock()
	delete(f.connMap, connId)
	atomic.AddInt64(&f.connCount, -1)

	//gc memory
	if f.connCount <= 0 {
		log.Printf("bucket %v, gc opt\n", f.bucketId)
		atomic.StoreInt64(&f.connCount, 0)
		runtime.GC()
	}

	return nil
}

//free run memory
func (f *Bucket) freeRunMemory() {
	//free memory
	f.Lock()
	defer f.Unlock()
	f.connMap = map[int64]iface.IConnect{}
	runtime.GC()
}

//init read message ticker
func (f *Bucket) initReadMsgTicker() {
	//get connect read msg rate
	readMsgRate := define.DefaultBucketReadRate

	//init read msg ticker
	f.readMsgTicker = queue.NewTicker(readMsgRate)
	f.readMsgTicker.SetCheckerCallback(f.cbForReadConnData)
}

//init send message queue and consumer
func (f *Bucket) initSendMsgConsumer() {
	//get send msg rate
	sendMsgRate := define.DefaultBucketSendRate
	f.sendMsgQueue = queue.NewList()
	f.sendMsgQueue.SetConsumer(f.cbForConsumerSendData, sendMsgRate)
}

//inter init
func (f *Bucket) interInit() {
	//init read msg ticker
	f.initReadMsgTicker()

	//init send message queue
	f.initSendMsgConsumer()
}
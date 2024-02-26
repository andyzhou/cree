package face

import (
	"errors"
	"github.com/andyzhou/cree/iface"
	"github.com/andyzhou/tinylib/queue"
	"sync"
)

/*
 * connect bucket face
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 *
 * - batch buckets contain all tcp connects
 * - one son worker contain batch connect
 */

//global variables
var (
	_bucket *Bucket
	_bucketOnce sync.Once
)

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
	worker *queue.Worker
	cbForReadMessage func(int64, iface.IRequest) error
	sync.RWMutex
}

//init
func init() {
	GetBucket()
}

//get single instance
func GetBucket() *Bucket {
	_bucketOnce.Do(func() {
		_bucket = NewBucket()
	})
	return _bucket
}

//construct
func NewBucket() *Bucket {
	this := &Bucket{
		worker: queue.NewWorker(),
	}
	this.interInit()
	return this
}

//quit
func (f *Bucket) Quit() {
	f.worker.Quit()
}

//////////////////
//api for cb func
//////////////////

//set cb for read message, step-1
func (f *Bucket) SetCBForReadMessage(cb func(int64, iface.IRequest) error)  {
	if cb == nil {
		return
	}
	f.cbForReadMessage = cb
}

//set cb for bind connects read opt, step-2
func (f *Bucket) SetCBForReadOpt(cb func(int32, ...interface{}) error) {
	if cb == nil {
		return
	}
	f.worker.SetCBForBindObjTickerOpt(cb)
}

//create batch son workers, step-3
func (f *Bucket) CreateSonWorkers(num int, tickRates ...float64) error {
	//check
	if num <= 0 {
		return errors.New("invalid parameter")
	}
	//create batch son workers
	err := f.worker.CreateWorkers(num, tickRates...)
	return err
}

///////////////////////
//api for send message
///////////////////////

//send sync message
func (f *Bucket) SendMessage(data interface{}, connectIds ...int64) error {
	//check
	if data == nil || connectIds == nil {
		return errors.New("invalid parameter")
	}

	//send to target workers
	_, err := f.worker.SendData(data, connectIds)
	return err
}

func (f *Bucket) SendMessageToWorker(data interface{}, workerId int32) error {
	//check
	if data == nil || workerId <= 0 {
		return errors.New("invalid parameter")
	}

	//get target worker
	sonWorker, err := f.worker.GetWorker(workerId)
	if err != nil || sonWorker == nil {
		return err
	}

	//send to target worker
	_, err = sonWorker.SendData(data)
	return err
}

//cast async message to all son workers
func (f *Bucket) CastMessage(data interface{}) error {
	//check
	if data == nil {
		return errors.New("invalid parameter")
	}

	//cast to all
	err := f.worker.CastData(data)
	return err
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

	//get target son worker
	sonWorker, err := f.worker.GetTargetWorker(connId)
	if err != nil {
		return nil, err
	}
	if sonWorker == nil {
		return nil, errors.New("can't get son worker")
	}

	//get connect
	obj, subErr := sonWorker.GetBindObj(connId)
	if subErr != nil || obj == nil {
		return nil, subErr
	}
	conn, ok := obj.(iface.IConnect)
	if !ok || conn == nil {
		return nil, errors.New("invalid obj type")
	}
	return conn, nil
}

//remove connect
func (f *Bucket) RemoveConnect(connId int64) error {
	//check
	if connId <= 0 {
		return errors.New("invalid parameter")
	}

	//get target son worker
	sonWorker, err := f.worker.GetTargetWorker(connId)
	if err != nil {
		return err
	}
	if sonWorker == nil {
		return errors.New("can't get son worker")
	}

	//remove connect from target son worker
	err = sonWorker.RemoveBindObj(connId)
	return err
}

//add new connect
func (f *Bucket) AddConnect(conn iface.IConnect) error {
	//check
	if conn == nil {
		return errors.New("invalid parameter")
	}

	//get target son worker
	connId := int64(conn.GetConnId())
	sonWorker, err := f.worker.GetTargetWorker(connId)
	if err != nil {
		return err
	}
	if sonWorker == nil {
		return errors.New("can't get son worker")
	}

	//save new connect into target son worker
	err = sonWorker.UpdateBindObj(connId, conn)
	return err
}

///////////////
//private func
///////////////

//cb for read connect data
func (f *Bucket) cbForReadConnData(
	workerId int32,
	connMaps ...interface{}) error {
	var (
		req iface.IRequest
		err error
	)
	//check
	if workerId <= 0 || connMaps == nil ||
		len(connMaps) <= 0 {
		return errors.New("invalid parameter")
	}
	mapVal := connMaps[0]
	if mapVal == nil {
		return errors.New("invalid conn map data")
	}
	connMap, ok := mapVal.(map[int64]interface{})
	if !ok || connMap == nil || len(connMap) <= 0 {
		return errors.New("no any conn map data")
	}

	//loop read connect data
	for connId, conn := range connMap {
		//check connect
		if connId <= 0 || conn == nil {
			continue
		}
		connObj, subOk := conn.(iface.IConnect)
		if !subOk || connObj == nil {
			continue
		}

		//read message
		req, err = connObj.ReadMessage()
		if err != nil {
			//log.Printf("bucket.cbForReadConnData, read message failed, connId:%v, err:%v\n",
			//	connId, err.Error())
			connObj.Stop()
			continue
		}

		//check and call rad message cb
		if f.cbForReadMessage != nil {
			f.cbForReadMessage(connId, req)
		}
	}
	return err
}

//inter init
func (f *Bucket) interInit() {
	//set cb for read connect data
	f.worker.SetCBForBindObjTickerOpt(f.cbForReadConnData)
}
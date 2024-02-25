package face

import (
	"errors"
	"sync"
	"time"
)

/*
 * worker face
 */

//inter type
type (
	SonWorker struct {
		id int
		queue *Queue
		cbForWorker func(interface{}) (interface{}, error)
	}
)

//face info
type Worker struct {
	cbForWorker func(interface{}) (interface{}, error)
	workerMap map[int]*SonWorker
	workerId int
	sync.RWMutex
}

//construct
func NewWorker() *Worker {
	this := &Worker{
		workerMap: map[int]*SonWorker{},
	}
	return this
}

//quit
func (f *Worker) Quit() {
	f.Lock()
	defer f.Unlock()
	for k, v := range f.workerMap {
		v.Quit()
		delete(f.workerMap, k)
	}
}

//set cb for worker opt, STEP-1
func (f *Worker) SetCBForWorker(cb func(interface{}) (interface{}, error)) {
	if cb == nil {
		return
	}
	f.cbForWorker = cb
}

//create batch workers, STEP-2
func (f *Worker) CreateWorkers(num int) error {
	if num <= 0 {
		return errors.New("invalid parameter")
	}
	f.Lock()
	defer f.Unlock()
	for i := 0; i < num; i++ {
		f.workerId++
		sw := NewSonWorker(f.workerId, f.cbForWorker)
		f.workerMap[f.workerId] = sw
	}
	return nil
}

//send data to target son worker
func (f *Worker) SendToWorker(data interface{}, ids ...int64) error {
	var (
		id int64
		hashIdx int
	)
	//check
	if data == nil {
		return errors.New("invalid parameter")
	}
	if ids != nil && len(ids) > 0 {
		id = ids[0]
	}

	if id > 0 {
		//hashed by id
		hashIdx = int(id % int64(f.workerId))
	}else{
		//hashed by timestamp
		hashIdx = int(time.Now().UnixNano() % int64(f.workerId))
	}

	//get son worker
	f.Lock()
	defer f.Unlock()
	sw, ok := f.workerMap[hashIdx]
	if !ok || sw == nil {
		sw, _ = f.workerMap[0]
	}
	if sw == nil {
		return errors.New("can't get son worker")
	}

	//send to son worker queue
	err := sw.SendData(data)
	return err
}

/////////////////////
//api for son worker
/////////////////////

//construct
func NewSonWorker(id int, cb func(interface{}) (interface{}, error)) *SonWorker {
	//self init
	this := &SonWorker{
		id: id,
		cbForWorker: cb,
		queue: NewQueue(),
	}
	//set cb for queue
	this.queue.SetCallback(cb)
	return this
}

//son quit
func (f *SonWorker) Quit() {
	f.queue.Quit()
}

//send data to queue
func (f *SonWorker) SendData(data interface{}) error {
	//check
	if data == nil {
		return errors.New("invalid parameter")
	}
	if f.queue == nil {
		return errors.New("inter queue is nil")
	}
	//send to queue
	_, err := f.queue.SendData(data)
	return err
}
package face

import (
	"github.com/andyzhou/cree/iface"
	"errors"
	"sync"
	"fmt"
	"math/rand"
	"time"
	"log"
)

/*
 * face for message handler
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

 //macro define
 const (
 	//for queue
 	HandlerQueueSizeDefault = 5
 	HandlerQueueSizeMax = 128

 	//for chan
 	HandlerQueueChanSize = 128
 )

 //face info
 type Handler struct {
 	queueSize int
 	handlerMap map[uint32]iface.IRouter
 	handlerQueue map[int]*HandlerWorker
 	sync.RWMutex
 }

 //son worker
 type HandlerWorker struct {
 	handler iface.IHandler //parent handler
 	queueChan chan iface.IRequest
 	closeChan chan bool
 }
 
 //construct
func NewHandler() *Handler {
	//self init
	this := &Handler{
		queueSize:HandlerQueueSizeDefault,
		handlerMap:make(map[uint32]iface.IRouter),
		handlerQueue:make(map[int]*HandlerWorker),
	}

	//inter init
	this.interInit()

	return this
}

func NewHandlerWorker(handler iface.IHandler) *HandlerWorker {
	//self init
	this := &HandlerWorker{
		handler:handler,
		queueChan:make(chan iface.IRequest, HandlerQueueChanSize),
		closeChan:make(chan bool),
	}

	//spawn main process
	go this.runMainProcess()

	return this
}

/////////
//api
/////////

//set and modify queue size
func (f *Handler) SetQueueSize(queueSize int) {
	if queueSize <= 0 {
		return
	}
	if queueSize > HandlerQueueSizeMax {
		queueSize = HandlerQueueSizeMax
	}
	f.reSizeQueueSize(queueSize)
 }

//message handle in queue
func (f *Handler) SendToQueue(req iface.IRequest) {
	//get random worker
	worker := f.getRandomWorker()
	if worker == nil {
		return
	}

	//send to worker queue
	worker.queueChan <- req
}

//message handle
func (f *Handler) DoMessageHandle(req iface.IRequest) error {
	//get relate handler by message id
	messageId := req.GetMessage().GetId()
	handler, ok := f.handlerMap[messageId]
	if !ok {
		tips := fmt.Sprintf("no handler for message id:%d", messageId)
		log.Println("Handler::DoMessageHandle ", tips)
		return errors.New(tips)
	}

	//call relate handle
	handler.PreHandle(req)
	handler.Handle(req)
	handler.PostHandle(req)

	return nil
}


//add router
func (f *Handler) AddRouter(messageId uint32, router iface.IRouter) error {
	//basic check
	if messageId <= 0 || router == nil {
		return errors.New("invalid parameter")
	}

	//check
	_, ok := f.handlerMap[messageId]
	if ok {
		return nil
	}

	//add into map with locker
	f.Lock()
	defer f.Unlock()
	f.handlerMap[messageId] = router

	return nil
}

///////////////
//private func
///////////////

//run son worker main process
func (sh *HandlerWorker) runMainProcess() {
	var (
		req iface.IRequest
		isOk, needQuit bool
	)
	for {
		if needQuit && len(sh.queueChan) <= 0 {
			break
		}
		select {
		case req, isOk = <- sh.queueChan:
			if isOk {
				sh.handler.DoMessageHandle(req)
			}
		case <- sh.closeChan:
			needQuit = true
		}
	}
}

//resize handler queue
func (f *Handler) reSizeQueueSize(queueSize int) {
	if f.queueSize >= queueSize {
		//do nothing
		return
	}

	//dynamic create new handler worker
	f.Lock()
	defer f.Unlock()
	for i := f.queueSize; i <= queueSize; i++ {
		worker := NewHandlerWorker(f)
		f.handlerQueue[i] = worker
	}

	//update running queue size
	f.queueSize = queueSize
}

//get random worker
func (f *Handler) getRandomWorker() *HandlerWorker {
	//basic check
	if f.handlerQueue == nil {
		return nil
	}
	queueSize := len(f.handlerQueue)
	if queueSize <= 0 {
		return nil
	}

	//get random index
	randomIndex := f.getRandomVal(queueSize) + 1

	//get relate worker
	worker, ok := f.handlerQueue[randomIndex]
	if !ok {
		return nil
	}

	return worker
}

//get rand number
func (f *Handler) getRandomVal(maxVal int) int {
	randSand := rand.NewSource(time.Now().UnixNano())
	r := rand.New(randSand)
	return r.Intn(maxVal)
}

//inter init
func (f *Handler) interInit() {
	//init worker pool
	f.Lock()
	defer f.Unlock()
	for i := 1; i <= f.queueSize; i++ {
		worker := NewHandlerWorker(f)
		f.handlerQueue[i] = worker
	}
}
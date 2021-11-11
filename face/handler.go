package face

import (
	"errors"
	"fmt"
	"github.com/andyzhou/cree/iface"
	"log"
	"math/rand"
	"sync"
	"time"
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
 	HandlerQueueChanSize = 1024
 )

 //face info
 type Handler struct {
 	queueSize int
 	redirectRouter iface.IRouter
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
		closeChan:make(chan bool, 1),
	}
	//spawn main process
	go this.runMainProcess()
	return this
}

/////////
//api
/////////

//handler worker quit
func (sh *HandlerWorker) Quit() {
	defer func() {
		if err := recover(); err != nil {
			log.Println("HandlerWorker:Quit panic, err:", err)
		}
	}()
	sh.closeChan <- true
}

//handler quit
func (f *Handler) Quit() {
	if f.handlerQueue != nil {
		for _, hq := range f.handlerQueue {
			hq.Quit()
		}
	}
}

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

	//try catch panic
	defer func() {
		if err := recover(); err != nil {
			log.Println("Handler::SendToQueue panic happened, err:", err)
			return
		}
	}()

	//async send to worker queue
	select {
	case worker.queueChan <- req:
	}
}

//message handle
func (f *Handler) DoMessageHandle(req iface.IRequest) error {
	//get relate handler by message id
	messageId := req.GetMessage().GetId()
	handler, ok := f.handlerMap[messageId]
	if !ok {
		//check redirect router
		if f.redirectRouter != nil {
			//call relate handle
			f.redirectRouter.PreHandle(req)
			f.redirectRouter.Handle(req)
			f.redirectRouter.PostHandle(req)
			return nil
		}

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

//register redirect for unsupported message id
//used for all requests redirect to handler
func (f *Handler) RegisterRedirect(router iface.IRouter) error {
	if router == nil {
		return errors.New("invalid parameter")
	}

	//set router
	f.redirectRouter = router
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

	defer func() {
		if err := recover(); err != nil {
			log.Println("HandlerWorker:mainProcess panic, err:", err)
		}
		//close chan
		close(sh.queueChan)
		close(sh.closeChan)
	}()

	//loop
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
	for i := 1; i <= f.queueSize; i++ {
		worker := NewHandlerWorker(f)
		f.handlerQueue[i] = worker
	}
}
package face

import (
	"errors"
	"fmt"
	"github.com/andyzhou/cree/define"
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

 //face info
 type Handler struct {
 	redirectRouter iface.IRouter
 	handlerMap sync.Map //msgId -> iRouter
 	handlerQueue sync.Map //workerId -> iWorker
    queueSize int
 	sync.RWMutex
 }
 
 //construct
func NewHandler() *Handler {
	//self init
	this := &Handler{
		queueSize:define.HandlerQueueSizeDefault,
		handlerMap:sync.Map{},
		handlerQueue:sync.Map{},
	}
	//inter init
	this.interInit()
	return this
}

/////////
//api
/////////

//handler quit
func (f *Handler) Quit() {
	sf := func(k, v interface{}) bool {
		hq, ok := v.(*HandlerWorker)
		if ok && hq != nil {
			hq.Quit()
		}
		return true
	}
	f.handlerQueue.Range(sf)
	f.handlerQueue = sync.Map{}
}

//set and modify queue size
func (f *Handler) SetQueueSize(queueSize int) {
	if queueSize <= 0 {
		return
	}
	if queueSize > define.HandlerQueueSizeMax {
		queueSize = define.HandlerQueueSizeMax
	}
	f.reSizeQueueSize(queueSize)
 }

//message handle in queue
func (f *Handler) SendToQueue(req iface.IRequest) {
	var (
		m any = nil
	)
	//get random worker
	worker := f.getRandomWorker()
	if worker == nil {
		return
	}

	//try catch panic
	defer func() {
		if err := recover(); err != m {
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
	router := f.getRouter(messageId)
	if router == nil {
		//check redirect router
		if f.redirectRouter != nil {
			//call relate handle
			f.redirectRouter.PreHandle(req)
			f.redirectRouter.Handle(req)
			f.redirectRouter.PostHandle(req)
			return nil
		}
		tips := fmt.Sprintf("no handler for message id:%d", messageId)
		log.Println("Handler::DoMessageHandle, tips:", tips)
		return errors.New(tips)
	}

	//call relate handle
	router.PreHandle(req)
	router.Handle(req)
	router.PostHandle(req)
	return nil
}

//add router
func (f *Handler) AddRouter(messageId uint32, router iface.IRouter) error {
	//basic check
	if messageId <= 0 || router == nil {
		return errors.New("invalid parameter")
	}

	//check
	oldRouter := f.getRouter(messageId)
	if oldRouter != nil {
		return nil
	}

	//add into map
	f.handlerMap.Store(messageId, router)
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

//resize handler queue
func (f *Handler) reSizeQueueSize(queueSize int) {
	if f.queueSize >= queueSize {
		//do nothing
		return
	}

	//dynamic create new handler worker
	for i := f.queueSize; i <= queueSize; i++ {
		worker := NewHandlerWorker(i, f)
		f.handlerQueue.Store(i, worker)
	}
}

//get random worker
func (f *Handler) getRandomWorker() *HandlerWorker {
	//basic check
	if f.queueSize <= 0 {
		return nil
	}
	//get random index
	randomIndex := f.getRandomVal(f.queueSize) + 1

	//get relate worker
	worker := f.getWorker(randomIndex)
	return worker
}

//get rand number
func (f *Handler) getRandomVal(maxVal int) int {
	randSand := rand.NewSource(time.Now().UnixNano())
	r := rand.New(randSand)
	return r.Intn(maxVal)
}

//get worker
func (f *Handler) getWorker(idx int) *HandlerWorker {
	v, ok := f.handlerQueue.Load(idx)
	if !ok || v == nil {
		return nil
	}
	worker, subOk := v.(*HandlerWorker)
	if !subOk || worker == nil {
		return nil
	}
	return worker
}

//get router of message id
func (f *Handler) getRouter(msgId uint32) iface.IRouter {
	if msgId < 0 {
		return nil
	}
	v, ok := f.handlerMap.Load(msgId)
	if !ok || v == nil {
		return nil
	}
	router, subOk := v.(iface.IRouter)
	if !subOk || router == nil {
		return nil
	}
	return router
}

//inter init
func (f *Handler) interInit() {
	//init worker pool
	for i := 1; i <= f.queueSize; i++ {
		worker := NewHandlerWorker(i, f)
		f.handlerQueue.Store(i, worker)
	}
}
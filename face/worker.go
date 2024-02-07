package face

import (
	"github.com/andyzhou/cree/define"
	"github.com/andyzhou/cree/iface"
	"log"
)

/*
 * face for handler son worker
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//son worker face
type HandlerWorker struct {
	idx int
	handler iface.IHandler //parent handler reference
	queueChan chan iface.IRequest
	closeChan chan bool
}

//construct
func NewHandlerWorker(idx int, handler iface.IHandler) *HandlerWorker {
	//self init
	this := &HandlerWorker{
		idx: idx,
		handler:handler,
		queueChan:make(chan iface.IRequest, define.HandlerQueueChanSize),
		closeChan:make(chan bool, 1),
	}
	//spawn main process
	go this.runMainProcess()
	return this
}

//handler worker quit
func (f *HandlerWorker) Quit() {
	var (
		m any = nil
	)
	defer func() {
		if subErr := recover(); subErr != m {
			log.Println("HandlerWorker:Quit panic, err:", subErr)
		}
	}()
	if f.closeChan != nil {
		f.closeChan <- true
	}
}

//run son worker main process
func (f *HandlerWorker) runMainProcess() {
	var (
		req iface.IRequest
		isOk bool
		m any = nil
	)

	defer func() {
		if err := recover(); err != m {
			log.Println("HandlerWorker:mainProcess panic, err:", err)
		}
		//process left queue
		for {
			req, isOk = <- f.queueChan
			if !isOk || req == nil {
				break
			}
			f.handler.DoMessageHandle(req)
		}
		close(f.queueChan)
	}()

	//loop
	for {
		select {
		case req, isOk = <- f.queueChan:
			if isOk && &req != nil {
				//call handler to process message request
				f.handler.DoMessageHandle(req)
			}
		case <- f.closeChan:
			return
		}
	}
}

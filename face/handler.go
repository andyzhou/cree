package face

import (
	"errors"
	"fmt"
	"sync"

	"github.com/andyzhou/cree/iface"
)

/*
 * face for message handler
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//face info
type Handler struct {
	redirectRouter iface.IRouter
	handlerMap     map[uint32]iface.IRouter //msgId -> iRouter
	sync.RWMutex
}

//construct
func NewHandler() *Handler {
	//self init
	this := &Handler{
		handlerMap: map[uint32]iface.IRouter{},
	}
	return this
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
	f.Lock()
	defer f.Unlock()
	f.handlerMap[messageId] = router
	return nil
}

//register redirect for unsupported message id
//used for all requests redirect to handler
func (f *Handler) RegisterRedirect(router iface.IRouter) error {
	//check
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

//get router of message id
func (f *Handler) getRouter(msgId uint32) iface.IRouter {
	if msgId < 0 {
		return nil
	}
	f.Lock()
	defer f.Unlock()
	v, ok := f.handlerMap[msgId]
	if !ok || v == nil {
		return nil
	}
	return v
}
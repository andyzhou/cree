package iface

/*
 * server interface
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

 //interface of server
 type IServer interface {
 	Start()
 	Stop()
	AddRouter(uint32,IRouter)
 	RegisterRedirect(IRouter)
 	GetManager()IManager
 	//setting
 	SetMaxConnects(int)
 	SetHandlerQueues(int)
 	//set hook
 	SetOnConnStart(func(IConnect))
 	SetOnConnStop(func(IConnect))
 	//callback
 	CallOnConnStart(IConnect)
 	CallOnConnStop(IConnect)
 }
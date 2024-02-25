package iface

/*
 * server interface
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

 //interface of server
 type IServer interface {
	//general
 	Start()
 	Stop()
	GetManager()IManager
	GetPacket()IPacket

	//setup
	AddRouter(uint32,IRouter)
 	RegisterRedirect(IRouter)

 	//setting
 	SetMaxConnects(int32)
 	//SetHandlerQueues(int)

 	//set hook
 	SetOnConnStart(func(IConnect))
 	SetOnConnStop(func(IConnect))

 	//callback
 	CallOnConnStart(IConnect)
 	CallOnConnStop(IConnect)
 }
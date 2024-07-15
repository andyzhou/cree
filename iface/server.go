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
	GetPacket()IPacket

	//setup
	AddRouter(uint32,IRouter)
 	RegisterRedirect(IRouter)

 	//setting
 	SetMaxConnects(int32)

 	//set hook
 	SetConnected(func(IConnect))
	SetDisconnected(func(IConnect))
	SetGenConnId(func()int64)
 }
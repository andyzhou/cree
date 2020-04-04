package iface

/*
 * interface for message handler
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

 type IHandler interface {
 	SetQueueSize(int)
 	DoMessageHandle(IRequest) error
 	AddRouter(uint32,IRouter) error
 	SendToQueue(IRequest)
 }

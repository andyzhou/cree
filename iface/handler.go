package iface

/*
 * interface for message handler
 */

 type IHandler interface {
 	SetQueueSize(int)
 	DoMessageHandle(IRequest) error
 	AddRouter(uint32,IRouter) error
 	SendToQueue(IRequest)
 }

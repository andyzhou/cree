package iface

/*
 * interface for request router
 */

 //interface info
 type IRouter interface {
 	PreHandle(IRequest)
 	Handle(IRequest)
 	PostHandle(IRequest)
 }

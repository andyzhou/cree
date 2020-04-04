package iface

/*
 * interface for request router
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

 //interface info
 type IRouter interface {
 	PreHandle(IRequest)
 	Handle(IRequest)
 	PostHandle(IRequest)
 }

package iface

/*
 * interface for message handler
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

 type IHandler interface {
 	DoMessageHandle(IRequest) error
 	AddRouter(uint32,IRouter) error
 	RegisterRedirect(IRouter) error
 }

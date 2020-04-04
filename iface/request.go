package iface

/*
 * interface for request
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

 type IRequest interface {
 	GetConnect() IConnect
 	GetMessage() IMessage
 }
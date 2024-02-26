package iface

import "net"

/*
 * interface for connect
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

 type IConnect interface {
	//base
 	Start()
 	Stop()
 	SendMessage(uint32, []byte) error
	ReadMessage() (IRequest, error)

	//get base
	GetActiveTime() int64
 	GetConn() *net.TCPConn
 	GetConnId() uint32
 	GetRemoteAddr() net.Addr

	//for property
 	RemoveProperty(string)
 	SetProperty(string,interface{}) bool
 	GetProperty(string)(interface{},error)
 }

package iface

import "net"

/*
 * interface for connect
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

 type IConnect interface {
 	Start()
 	Stop()
 	SendMessage(uint32, []byte) error
 	GetConn() *net.TCPConn
 	GetConnId() uint32
 	GetRemoteAddr() net.Addr
 	RemoveProperty(string)
 	SetProperty(string,interface{}) bool
 	GetProperty(string)(interface{},error)
 }

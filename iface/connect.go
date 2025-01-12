package iface

import "net"

/*
 * interface for connect
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

 type IConnect interface {
	//base
	Quit()
 	SendMessage(uint32, []byte) error
	ReadMessage() (IRequest, error)

	//get base
	GetActiveTime() int64
 	GetConn() *net.TCPConn
 	GetConnId() int64
 	GetRemoteAddr() net.Addr

	//for tag
	RemoveTags(tags ...string) error
	GetTags() map[string]bool
	SetTag(tags ...string) error

	//for property
	GetGroupId() int64
	SetGroupId(groupId int64) error
 	RemoveProperty(string)
 	GetProperty(string)(interface{},error)
	SetProperty(string,interface{}) bool
 }

package iface

import "github.com/andyzhou/cree/define"

/*
 * interface for bucket
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

type IBucket interface {
	//gen opt
	Quit()

	//conn opt
	GetConnect(connId int64) (IConnect, error)
	RemoveConnect(connId int64) error
	AddConnect(conn IConnect) error

	//msg opt
	SendMessage(req *define.SendMsgReq) error

	//set cb
	SetCBForReadMessage(cb func(IConnect, IRequest) error)
	SetCBForDisconnected(cb func(IConnect))
}

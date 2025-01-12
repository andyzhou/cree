package iface


/*
 * interface for group
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

type IGroup interface {
	Clear()
	SendMessage(msgId uint32, msg []byte) error
	Quit(connections ...IConnect) error
	Join(conn IConnect) error
	SetErrMsgId(msgId uint32)
}

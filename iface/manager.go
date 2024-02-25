package iface

/*
 * interface for manager
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

 type IManager interface {
	Quit()
	Clear()
 	Add(IConnect) error
 	Remove(IConnect) error
 	Get(uint32)(IConnect,error)
 	GetLen()int32
	SetUnActiveSeconds(val int)
 }

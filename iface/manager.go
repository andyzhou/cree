package iface

/*
 * interface for manager
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

 type IManager interface {
	Quit()
 	Add(IConnect) error
 	Remove(IConnect) error
 	Clear()
 	Get(uint32)(IConnect,error)
 	GetLen()int32
	SetUnActiveSeconds(val int)
 }

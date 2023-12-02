package iface

/*
 * interface for manager
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

 type IManager interface {
 	Add(IConnect) error
 	Remove(IConnect) error
 	Clear()
 	Get(uint32)(IConnect,error)
 	GetLen()int32
 }

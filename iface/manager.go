package iface

/*
 * interface for manager
 */

 type IManager interface {
 	Add(IConnect)
 	Remove(IConnect)
 	Clear()
 	Get(uint32)(IConnect,error)
 	GetLen()int
 }

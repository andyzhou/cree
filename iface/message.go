package iface

/*
 * interface for message
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

 type IMessage interface {
 	//get
 	GetLen()uint32
 	GetKind()uint32
 	GetId()uint32
 	GetData()[]byte

 	//set
 	SetKind(uint32)
 	SetId(uint32)
 	SetData([]byte)
 	SetLen(uint32)
 }

package iface

/*
 * interface for message
 */

 type IMessage interface {
 	//get
 	GetLen()uint32
 	GetId()uint32
 	GetData()[]byte

 	//set
 	SetId(uint32)
 	SetData([]byte)
 	SetLen(uint32)
 }

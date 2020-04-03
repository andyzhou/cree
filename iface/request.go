package iface

/*
 * interface for request
 */

 type IRequest interface {
 	GetConnect() IConnect
 	GetMessage() IMessage
 }
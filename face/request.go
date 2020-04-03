package face

import "github.com/andyzhou/cree/iface"

/*
 * face for request
 */

 //face info
 type Request struct {
 	conn iface.IConnect //connect for client
 	message iface.IMessage //message from client
 }

 //construct
func NewRequest(conn iface.IConnect, message iface.IMessage) *Request {
	this := &Request{
		conn:conn,
		message:message,
	}
	return this
}

////////
//api
///////

func (r *Request) GetMessage() iface.IMessage {
	return r.message
}

func (r *Request) GetConnect() iface.IConnect {
	return r.conn
}


package face

import (
	"github.com/andyzhou/cree/iface"
)

/*
 * face for router
 */

 //face info
 type BaseRouter struct {
 }

func (br *BaseRouter)PreHandle(req iface.IRequest){}
func (br *BaseRouter)Handle(req iface.IRequest){}
func (br *BaseRouter)PostHandle(req iface.IRequest){}


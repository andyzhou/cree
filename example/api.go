package main

import (
	"github.com/andyzhou/cree/face"
	"github.com/andyzhou/cree/iface"
	"fmt"
)

/*
 * Testing api
 */

 type TestApi struct {
 	face.BaseRouter
 }

func (*TestApi) Handle(req iface.IRequest) {
	fmt.Println("TestApi::Handle, data:", string(req.GetMessage().GetData()))
	message := req.GetMessage()
	req.GetConnect().SendMessage(message.GetId(), message.GetData())
}
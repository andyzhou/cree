package main

import (
	"fmt"
	"github.com/andyzhou/cree"
	"log"
	"os"
	"sync"
	"time"
)

/*
 * client testing
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//cb for client read
func CBForRead(data []byte) bool {
	log.Println("client:CBForRead, data:", string(data))
	return true
}

//test write
func ClientWrite(client *cree.Client) {
	//send packet data
	messageId := uint32(1)
	data := fmt.Sprintf("time:%d", time.Now().Unix())
	err := client.SendPacket(messageId, []byte(data), true)
	if err != nil {
		log.Println("ClientWrite failed, err:", err.Error())
	}
}

//main
func main() {
	var (
		wg sync.WaitGroup
		host = "127.0.0.1"
		port = 7800
	)

	//wg
	wg.Add(1)

	//init new client
	client := cree.NewClient(host, port)

	//setup
	client.SetCBForRead(CBForRead)

	//try connect server
	err := client.ConnServer()
	if err != nil {
		log.Println(err)
		wg.Done()
		os.Exit(1)
	}

	//spawn write testing
	go ClientWrite(client)

	wg.Wait()
}
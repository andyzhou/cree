package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/andyzhou/cree"
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
func ClientWrite(
	client *cree.Client,
	testTimes int,
	wg *sync.WaitGroup) {
	messageId := uint32(1)
	maxTimes := 100
	times := 1
	for {
		//send packet data
		data := fmt.Sprintf("time:%d", time.Now().Unix())
		err := client.SendPacket(messageId, []byte(data))
		if err != nil {
			log.Println("ClientWrite failed, err:", err.Error())
		}
		time.Sleep(time.Second)
		times++
		if times >= maxTimes {
			break
		}
	}
	wg.Done()
}

//main
func main() {
	var (
		wg sync.WaitGroup
		host = "127.0.0.1"
		port = 7800
		testTimes = 50
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
	log.Println("client start")
	go ClientWrite(client, testTimes, &wg)

	wg.Wait()
	log.Println("client closed")
}
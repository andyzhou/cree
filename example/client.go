package main

import (
	"errors"
	"fmt"
	"github.com/andyzhou/cree/iface"
	"log"
	"math/rand"
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
func CBForRead(msg iface.IMessage) error {
	//check
	if msg == nil {
		return errors.New("invalid parameter")
	}
	log.Printf("client:CBForRead, msgId:%v, data:%v\n", msg.GetId(), string(msg.GetData()))
	return nil
}

//test write
func ClientWrite(
	client *cree.Client,
	testTimes int,
	wg *sync.WaitGroup) {
	messageId := uint32(0)
	times := 1
	maxMsgId := 4
	for {
		//setup message id
		messageId = uint32(rand.Intn(maxMsgId) + 1)
		//send packet data
		data := fmt.Sprintf("time:%d", time.Now().Unix())
		err := client.SendPacket(messageId, []byte(data))
		if err != nil {
			log.Println("ClientWrite failed, err:", err.Error())
		}
		time.Sleep(time.Second/10)
		times++
		if testTimes > 0 && times >= testTimes {
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
		testTimes = 0
		clients = 1
	)

	//wg
	wg.Add(1)

	//set client conf
	clientCfg := &cree.ClientConf{
		Host: host,
		Port: port,
	}

	for i := 0; i < clients; i++ {
		//init new client
		client := cree.NewClient(clientCfg)

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
		go ClientWrite(client, testTimes, &wg)
	}

	wg.Wait()
	log.Println("client closed")
}
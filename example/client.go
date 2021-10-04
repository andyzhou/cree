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

// //pack test data
//func packTestData(clientId int) []byte {
//	data := fmt.Sprintf("clientId:%d, time:%d", clientId, time.Now().Unix())
//	message := face.NewMessage()
//	message.Id = 1
//	message.SetData([]byte(data))
//	packet := face.NewPacket()
//	byteData, _ := packet.Pack(message)
//	return byteData
//}
//
//func CreateOneClient(address string, id int)  {
//	conn, err := net.Dial("tcp", address)
//	if err != nil {
//		fmt.Println("client start failed, err:", err.Error())
//		return
//	}
//	i := 1
//	for {
//		//init packet
//		packTestData := packTestData(id)
//
//		//send
//		_, err := conn.Write(packTestData)
//		if err != nil {
//			break
//		}
//
//		//read
//		buf := make([]byte, 512)
//		cnt, err := conn.Read(buf)
//		if err != nil {
//			fmt.Println("read buf error ")
//			return
//		}
//		fmt.Printf(" server call back : %s, cnt = %d\n", string(buf), cnt)
//		if i >= 50 {
//			//up to limit of testing
//			break
//		}
//		time.Sleep(time.Second * 1)
//		i++
//	}
//	conn.Close()
//}
//
//func ClientTest(host string, port, clients int) {
//	time.Sleep(time.Second * 3)
//	addr := fmt.Sprintf("%s:%d", host, port)
//	for i := 1; i <= clients; i++ {
//		go CreateOneClient(addr, i)
//	}
//}

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
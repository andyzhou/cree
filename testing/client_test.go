package testing

import (
	"fmt"
	"github.com/andyzhou/cree/iface"
	"sync"
	"testing"
	"time"

	"github.com/andyzhou/cree"
)

var (
	host = "127.0.0.1"
	port = 7800
	clientMap = map[int]*cree.Client{}
	locker = sync.RWMutex{}
	cc *cree.Client
	err error
)

//init
func init() {
	cc, err = connClient()
	if err != nil {
		panic(any(err))
	}
}

//cb for client read
func cbForRead(msg iface.IMessage) error {
	//log.Println("client:CBForRead, data:", string(data))
	return nil
}

//connect client
func connClient() (*cree.Client, error) {
	//set client conf
	clientCfg := &cree.ClientConf{
		Host: host,
		Port: port,
	}
	//init new client
	client := cree.NewClient(clientCfg)
	client.SetCBForRead(cbForRead)
	err = client.ConnServer()
	return client, err
}

//test connect
func TestConnect(t *testing.T) {
	c, subErr := connClient()
	if subErr != nil {
		t.Errorf("test connect failed, err;%v\n", subErr.Error())
		return
	}
	defer c.Close()
	t.Logf("test connect success\n")
}

//test write
func TestWrite(t *testing.T) {
	messageId := uint32(1)
	data := fmt.Sprintf("time:%d", time.Now().Unix())
	subErr := cc.SendPacket(messageId, []byte(data))
	t.Logf("test write, subErr:%v\n", subErr)
}

func BenchmarkWrite(b *testing.B) {
	//send packet data
	messageId := uint32(1)
	succeed := 0
	failed := 0
	for i := 0; i < b.N; i++ {
		data := fmt.Sprintf("time:%d", time.Now().Unix())
		subErr := cc.SendPacket(messageId, []byte(data))
		if subErr != nil {
			failed++
		}else{
			succeed++
		}
	}
	b.Logf("benchmark write, succeed:%v, failed:%v\n", succeed, failed)
}

func BenchmarkConnect(b *testing.B) {
	succeed := 0
	failed := 0
	locker.Lock()
	locker.Unlock()
	for i := 0; i < b.N; i++ {
		c, subErr := connClient()
		if subErr != nil {
			failed++
		}else{
			clientMap[i] = c
			succeed++
		}
	}
	b.Logf("benchmark connect, succeed:%v, failed:%v\n", succeed, failed)
	//close
	for idx, v := range clientMap {
		if v != nil {
			v.Close()
			v = nil
		}
		delete(clientMap, idx)
	}
	clientMap = map[int]*cree.Client{}
}
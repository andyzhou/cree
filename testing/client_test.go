package testing

import (
	"github.com/andyzhou/cree"
	"sync"
	"testing"
)

var (
	host = "127.0.0.1"
	port = 7800
	clientMap = map[int]*cree.Client{}
	locker = sync.RWMutex{}
)

//connect client
func connClient() (*cree.Client, error) {
	//set client conf
	clientCfg := &cree.ClientConf{
		Host: host,
		Port: port,
	}
	//init new client
	client := cree.NewClient(clientCfg)
	err := client.ConnServer()
	return client, err
}

//test connect
func TestConnect(t *testing.T) {
	c, err := connClient()
	if err != nil {
		t.Errorf("test connect failed, err;%v\n", err.Error())
		return
	}
	defer c.Close()
	t.Logf("test connect success\n")
}

func BenchmarkConnect(b *testing.B) {
	succeed := 0
	failed := 0
	locker.Lock()
	locker.Unlock()
	for i := 0; i < b.N; i++ {
		c, err := connClient()
		if err != nil {
			b.Errorf("benmark connect failed, err:%v\n", err.Error())
			failed++
			break
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
package face

import (
	"errors"
	"fmt"
	"github.com/andyzhou/cree/iface"
	"io"
	"log"
	"net"
	"sync"
)

/*
 * face for connect
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

 //inter macro define
 const (
 	WriteChanSize = 1024
 )

 //face info
 type Connect struct {
 	tcpServer iface.IServer //parent tcp server
 	packet iface.IPacket //parent packet interface
 	connId uint32
 	isClosed bool
 	conn *net.TCPConn //socket tcp connect
 	handler iface.IHandler
 	messageChan chan []byte
 	closeChan chan bool
 	propertyMap map[string]interface{}
 	sync.RWMutex
 }

 //construct
func NewConnect(
				server iface.IServer,
				conn *net.TCPConn,
				connectId uint32,
				handler iface.IHandler,
			) *Connect {
	//self init
	this := &Connect{
		tcpServer:server,
		packet: server.GetPacket(),
		conn:conn,
		connId:connectId,
		handler:handler,
		messageChan:make(chan []byte, WriteChanSize),
		closeChan:make(chan bool, 1),
		propertyMap:make(map[string]interface{}),
	}

	//spawn main process
	go this.runMainProcess()

	return this
}

//////////
//api
/////////

//send message
func (c *Connect) SendMessage(messageId uint32, data []byte) (err error) {
	var (
		m any = nil
	)
	//basic check
	if messageId <= 0 || data == nil {
		err = errors.New("invalid parameter")
		return
	}

	//create message
	message := NewMessage()
	message.SetId(messageId)
	message.SetData(data)

	//create message packet
	byteData, err := c.packet.Pack(message)
	if err != nil {
		return
	}

	//try catch panic
	defer func(result error) {
		if subErr := recover(); subErr != m {
			tips := fmt.Sprintln("panic happened, err:", subErr)
			result = errors.New(tips)
		}
	}(err)

	//send data to chan
	c.messageChan <- byteData

	return
}

//start
func (c *Connect) Start() {
	//start read and write goroutine
	go c.startRead()
	go c.startWrite()

	//call hook of connect start
	c.tcpServer.CallOnConnStart(c)
}

func (c *Connect) Stop() {
	var (
		m any = nil
	)
	if c.isClosed == true {
		return
	}

	defer func() {
		if err := recover(); err != m {
			log.Println("cree:Connect::Stop, panic happened, err:", err)
		}
		//close connect
		c.conn.Close()

		//close chan
		close(c.messageChan)
		close(c.closeChan)
	}()

	c.isClosed = true

	//call hook of connect closed
	c.tcpServer.CallOnConnStop(c)

	//close tcp
	c.closeChan <- true

	//remove from manager
	c.tcpServer.GetManager().Remove(c)
}

//get remote client address
func (c *Connect) GetRemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

//get connect
func (c *Connect) GetConn() *net.TCPConn {
	return c.conn
}

//get connect id
func (c *Connect) GetConnId() uint32 {
	return c.connId
}

//remove property
func (c *Connect) RemoveProperty(key string) {
	if key == "" {
		return
	}
	//remove property with locker
	c.Lock()
	defer c.Unlock()
	delete(c.propertyMap, key)
}

//set property
func (c *Connect) SetProperty(key string, value interface{}) bool {
	if key == "" || value == nil {
		return false
	}

	//sync property with locker
	c.Lock()
	defer c.Unlock()
	c.propertyMap[key] = value

	return true
}

//get property
func (c *Connect) GetProperty(key string) (interface{}, error) {
	v, ok := c.propertyMap[key]
	if !ok {
		return nil, errors.New("no property value")
	}
	return v, nil
}

///////////////
//private func
//////////////

//run main process
func (c *Connect) runMainProcess() {
	//call hook of connect start
	c.tcpServer.CallOnConnStart(c)

	//start read and write goroutine
	go c.startRead()
	go c.startWrite()
}

//read data from client
func (c *Connect) startRead() {
	var (
		data []byte
		message iface.IMessage
		err error
		m any = nil
	)

	//defer function
	defer func() {
		if subErr := recover(); subErr != m {
			log.Println("Connect:startRead panic, err:", subErr)
		}
		//stop connect
		c.Stop()
	}()

	//init header
	header := make([]byte, c.packet.GetHeadLen())

	//read data in the loop
	for {
		//read message head
		_, err = io.ReadFull(c.conn, header)
		if err != nil {
			log.Println("read message header failed, err:", err.Error())
			break
		}

		//unpack header
		message, err = c.packet.UnPack(header)
		if err != nil {
			log.Println("unpack message failed, err:", err.Error())
			break
		}

		//read real data and storage into message object
		if message.GetLen() > 0 {
			data = make([]byte, message.GetLen())
			_, err = io.ReadFull(c.conn, data)
			if err != nil {
				log.Println("read data failed, err:", err.Error())
				break
			}
			message.SetData(data)
		}

		//init client request
		req := NewRequest(c, message)

		//send request to handler queue
		c.handler.SendToQueue(req)
	}
}

//send data for client
func (c *Connect) startWrite() {
	var (
		data = make([]byte, 0)
		err error
		isOk, needQuit bool
		m any = nil
	)

	defer func() {
		if subErr := recover(); subErr != m {
			log.Println("Connect:startWrite panic, err:", subErr)
		}
		//stop connect
		c.Stop()
	}()

	//loop
	for {
		if needQuit && len(c.messageChan) <= 0 {
			break
		}
		select {
		case data, isOk = <- c.messageChan:
			{
				if isOk {
					if c.conn == nil {
						needQuit = true
						break
					}
					_, err = c.conn.Write(data)
					if err != nil {
						log.Println("send data to client failed, err:", err.Error())
						return
					}
				}
			}
		case <- c.closeChan:
			{
				return
			}
		}
	}
}

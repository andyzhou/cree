package face

import (
	"errors"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/andyzhou/cree/iface"
)

/*
 * face for connect
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

 //face info
type Connect struct {
	tcpServer   iface.IServer //parent tcp server reference
	packet      iface.IPacket //parent packet interface reference
	conn        *net.TCPConn //socket tcp connect
	handler     iface.IHandler
	propertyMap map[string]interface{}
	connId      uint32
	isClosed    bool
	activeTime  int64 //last active timestamp
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
		propertyMap:make(map[string]interface{}),
	}
	return this
}

//get last active time
func (c *Connect) GetActiveTime() int64 {
	return c.activeTime
}

//send message direct
func (c *Connect) SendMessage(
	messageId uint32, data []byte) error {
	//basic check
	if messageId <= 0 || data == nil {
		return errors.New("invalid parameter")
	}
	if c.conn == nil {
		return errors.New("connect is nil")
	}

	//create message
	message := NewMessage()
	message.SetId(messageId)
	message.SetData(data)

	//create message packet
	byteData, err := c.packet.Pack(message)
	if err != nil {
		return err
	}

	//defer update active time
	defer func() {
		c.activeTime = time.Now().Unix()
	}()

	//direct send with locker
	_, err = (*c.conn).Write(byteData)
	return err
}

//start
func (c *Connect) Start() {
	//start read and write goroutine
	go c.startRead()

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
		c.isClosed = true
	}()

	//call hook of connect closed
	c.tcpServer.CallOnConnStop(c)

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
	//check
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
	c.Lock()
	defer c.Unlock()
	v, ok := c.propertyMap[key]
	if !ok {
		return nil, errors.New("no property value")
	}
	return v, nil
}

///////////////
//private func
//////////////

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
			log.Println("cree.connect.startRead panic, err:", subErr)
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
			//log.Println("cree.connect, read message header failed, err:", err.Error())
			break
		}

		//unpack header
		message, err = c.packet.UnPack(header)
		if err != nil {
			log.Println("cree.connect.startRead, unpack message failed, err:", err.Error())
			break
		}

		//read real data and storage into message object
		if message.GetLen() > 0 {
			data = make([]byte, message.GetLen())
			_, err = io.ReadFull(c.conn, data)
			if err != nil {
				log.Println("cree.connect.startRead, read data failed, err:", err.Error())
				break
			}
			message.SetData(data)
		}

		//init client request
		req := NewRequest(c, message)

		//handle request message
		c.handler.DoMessageHandle(req)
	}
}

//cb for list consumer
func (c *Connect) cbForConsumer(data interface{}) (interface{}, error) {
	//check
	if data == nil {
		return nil, errors.New("invalid parameter")
	}
	dataBytes, ok := data.([]byte)
	if !ok || dataBytes == nil {
		return nil, errors.New("data should be `[]byte` type")
	}
	if c.conn == nil {
		return nil, errors.New("connect is nil")
	}
	//send to connect
	_, err := (*c.conn).Write(dataBytes)
	return nil, err
}
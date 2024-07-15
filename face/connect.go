package face

import (
	"errors"
	"fmt"
	"io"
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
	conn        *net.TCPConn  //socket tcp connect
	handler     iface.IHandler
	tagMap      map[string]bool
	propertyMap map[string]interface{}
	connId      int64
	isClosed    bool
	activeTime  int64 //last active timestamp
	sync.RWMutex
}

 //construct
func NewConnect(
		server iface.IServer,
		conn *net.TCPConn,
		connectId int64,
		handler iface.IHandler,
	) *Connect {
	//self init
	this := &Connect{
		tcpServer:server,
		packet: server.GetPacket(),
		conn:conn,
		connId:connectId,
		handler:handler,
		tagMap: map[string]bool{},
		propertyMap:make(map[string]interface{}),
	}
	return this
}

//quit
func (c *Connect) Quit() {
	c.Lock()
	defer c.Unlock()
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}

//get last active time
func (c *Connect) GetActiveTime() int64 {
	return c.activeTime
}

//send message direct
func (c *Connect) SendMessage(messageId uint32, data []byte) error {
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

//read message
func (c *Connect) ReadMessage() (iface.IRequest, error) {
	//init header
	header := make([]byte, c.packet.GetHeadLen())

	//read one message
	req, err := c.readOneMessage(header)
	return req, err
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
func (c *Connect) GetConnId() int64 {
	return c.connId
}

//remove tags
func (c *Connect) RemoveTags(tags ...string) error {
	//check
	if tags == nil || len(tags) <= 0 {
		return errors.New("invalid parameter")
	}

	//del mark with locker
	c.Lock()
	defer c.Unlock()
	for _, tag := range tags {
		delete(c.tagMap, tag)
	}
	return nil
}

//get tags
func (c *Connect) GetTags() map[string]bool {
	//get all tags
	c.Lock()
	defer c.Unlock()
	return c.tagMap
}

//set new tags
func (c *Connect) SetTag(tags ...string) error {
	//check
	if tags == nil || len(tags) <= 0 {
		return errors.New("invalid parameter")
	}

	//mark with locker
	c.Lock()
	defer c.Unlock()
	for _, tag := range tags {
		c.tagMap[tag] = true
	}
	return nil
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

////read data from client
//func (c *Connect) startRead() {
//	var (
//		err error
//		m any = nil
//	)
//
//	//defer function
//	defer func() {
//		if subErr := recover(); subErr != m {
//			log.Println("cree.connect.startRead panic, err:", subErr)
//		}
//	}()
//
//	//init header
//	header := make([]byte, c.packet.GetHeadLen())
//
//	//read data in the loop
//	for {
//		//read one message
//		_, err = c.readOneMessage(header)
//		if err != nil {
//			break
//		}
//	}
//}

//read one message
func (c *Connect) readOneMessage(header []byte) (iface.IRequest, error) {
	//check
	if header == nil {
		return nil, errors.New("invalid header parameter")
	}
	if c.conn == nil {
		return nil, errors.New("connect is nil")
	}

	//read message head
	_, err := io.ReadFull(c.conn, header)
	if err != nil {
		return nil, err
	}

	//unpack header
	message, subErr := c.packet.UnPack(header)
	if subErr != nil {
		errTip := fmt.Errorf("cree.connect.startRead, unpack message failed, err:%v", subErr.Error())
		return nil, errTip
	}

	//read real data and storage into message object
	if message.GetLen() > 0 {
		data := make([]byte, message.GetLen())
		_, err = io.ReadFull(c.conn, data)
		if err != nil {
			errTip := fmt.Errorf("cree.connect.startRead, read data failed, err:%v", err.Error())
			return nil, errTip
		}
		message.SetData(data)
	}

	//init client request
	req := NewRequest(c, message)

	//handle request message
	err = c.handler.DoMessageHandle(req)
	return req, err
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
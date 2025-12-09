package cree

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/andyzhou/cree/define"
	"github.com/andyzhou/cree/face"
	"github.com/andyzhou/cree/iface"
)

/*
 * face for client
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//client config
type ClientConf struct {
	Host         string
	Port         int
	ReadBuffSize int
	ConnTimeOut  time.Duration
	WriteTimeOut int //xx seconds
}

//face info
type Client struct {
	conf *ClientConf
	conn net.Conn
	connected bool
	cbForRead func(msg iface.IMessage) error
	pack iface.IPacket
	writeMu sync.Mutex
}

//construct
func NewClient(
		conf *ClientConf,
	) *Client {
	//self init
	this := &Client{
		conf: conf,
		pack: face.NewPacket(),
	}

	//inter init
	this.interInit()
	return this
}

//close connect
func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
	c.connected = false
}

//set cb for read data
func (c *Client) SetCBForRead(cb func(msg iface.IMessage) error) bool {
	if cb == nil {
		return false
	}
	c.cbForRead = cb
	return true
}

//set max pack size
func (c *Client) SetMaxPackSize(size int) {
	c.pack.SetMaxPackSize(size)
}

//set read buff size
func (c *Client) SetReadBuffSize(size int) bool {
	if size <= 0 {
		return false
	}
	c.conf.ReadBuffSize = size
	return true
}

//send packet data
func (c *Client) SendPacket(
	messageId uint32,
	data []byte) error {
	var (
		m any = nil
	)
	//check
	if data == nil {
		return errors.New("invalid parameter")
	}
	if !c.connected || c.conn == nil {
		return errors.New("connect is nil")
	}

	//try catch panic
	defer func() {
		if err := recover(); err != m {
			log.Printf("cree.client.SendPacket err:%v\n", err)
		}
	}()

	//packet data
	packet := c.packetData(messageId, data)

	//set write timeout
	writeTimeOut := time.Duration(c.conf.WriteTimeOut)  * time.Second

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	//send direct
	err := c.conn.SetWriteDeadline(time.Now().Add(writeTimeOut))
	if err != nil {
		return err
	}
	_, err = c.conn.Write(packet)
	c.conn.SetWriteDeadline(time.Time{}) // 清除 deadline
	if err != nil {
		log.Printf("cree.client.SendPacket failed err:%v\n", err)
	}
	return err
}

//connect server
func (c *Client) ConnServer(timeOuts ...time.Duration) error {
	var (
		timeOut time.Duration
		conn net.Conn
		err error
	)
	//check
	if c.conf.Host == "" || c.conf.Port <= 0 {
		return errors.New("host or port is invalid")
	}
	if c.connected {
		return nil
	}

	if timeOuts != nil && len(timeOuts) > 0 {
		timeOut = timeOuts[0]
	}else{
		timeOut = c.conf.ConnTimeOut
	}
	if timeOut <= 0 {
		timeOut = define.DefaultTcpDialTimeOut * time.Second
	}
	//log.Printf("confTimeout:%v, timeout:%v\n", c.conf.ConnTimeOut, timeOut)

	//format address
	address := fmt.Sprintf("%s:%d", c.conf.Host, c.conf.Port)

	//try connect server
	if timeOut > 0 {
		conn, err = net.DialTimeout("tcp", address, timeOut)
	}else{
		conn, err = net.Dial("tcp", address)
	}
	if err != nil {
		return err
	}

	//sync conn
	c.conn = conn
	c.connected = true

	//spawn read process
	//go c.runReadProcess()
	return nil
}

////////////////
//private func
////////////////

//packet one data
func (c *Client) packetData(
	messageId uint32,
	data []byte) []byte {
	message := face.NewMessage()
	message.Id = messageId
	message.SetData(data)
	byteData, _ := c.pack.Pack(message)
	return byteData
}

//read one message
func (c *Client) readMessage() (iface.IMessage, error) {
	//init header
	header := make([]byte, c.pack.GetHeadLen())

	//read message header
	_, err := io.ReadFull(c.conn, header)
	if err != nil {
		return nil, err
	}

	//unpack header
	message, subErr := c.pack.UnPack(header)
	if subErr != nil || message == nil {
		return nil, subErr
	}

	//read real data and storage into message object
	if message.GetLen() > 0 {
		data := make([]byte, message.GetLen())
		_, err = c.conn.Read(data)
		if err != nil {
			return nil, err
		}
		message.SetData(data)
	}
	return message, nil
}

//read process
func (c *Client) runReadProcess() {
	var (
		//buff = make([]byte, c.conf.ReadBuffSize)
		msg iface.IMessage
		err error
		m any = nil
	)

	//defer
	defer func() {
		if subErr := recover(); subErr != m {
			log.Println("client.runReadProcess panic, err:", err)
		}
		c.Close()
	}()

	//loop
	for {
		//check connect
		if c.conn == nil {
			continue
		}

		//try read tcp data
		msg, err = c.readMessage()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				break
			}
			continue
		}

		//call cb
		if c.cbForRead != nil {
			//unpack data
			c.cbForRead(msg)
		}
	}
}

//cb for consumer
func (c *Client) cbForConsumer(
	data interface{}) error {
	//check
	if data == nil {
		return errors.New("invalid parameter")
	}
	packBytes, ok := data.([]byte)
	if !ok || packBytes == nil {
		return errors.New("data should be `[]byte` type")
	}
	if c.conn == nil {
		return errors.New("client connect is nil")
	}
	//write to connect
	_, err := c.conn.Write(packBytes)
	return err
}

//inter init
func (c *Client) interInit() {
	//check config
	if c.conf.ReadBuffSize <= 0 {
		c.conf.ReadBuffSize = define.DefaultTcpReadBuffSize
	}
	if c.conf.ConnTimeOut <= 0 {
		c.conf.ConnTimeOut = time.Second * define.DefaultTcpDialTimeOut
	}
	if c.conf.WriteTimeOut <= 0 {
		c.conf.WriteTimeOut = define.DefaultTcpWriteTimeOut
	}

	//start read process
	go c.runReadProcess()
}
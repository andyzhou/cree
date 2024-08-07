package cree

import (
	"errors"
	"fmt"
	"log"
	"net"
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
	Host string
	Port int
	ReadBuffSize int
	ConnTimeOut time.Duration
	WriteTimeOut int //xx seconds
}

//son client
type SonClient struct {
	conn *net.Conn
	connected bool
	pack iface.IPacket
}

//face info
type Client struct {
	conf *ClientConf
	conn *net.Conn
	connected bool
	cbForRead func(data []byte) bool
	pack iface.IPacket
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
		(*c.conn).Close()
		c.conn = nil
	}
	c.connected = false
}

//set cb for read data
func (c *Client) SetCBForRead(
	cb func(data []byte) bool) bool {
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
	//check
	if messageId < 0 || data == nil {
		return errors.New("invalid parameter")
	}
	if c.conn == nil {
		return errors.New("connect is nil")
	}

	//packet data
	packet := c.packetData(messageId, data)

	//set write timeout
	writeTimeOut := time.Duration(c.conf.WriteTimeOut)  * time.Second

	//send direct
	err := (*c.conn).SetWriteDeadline(time.Now().Add(writeTimeOut))
	if err != nil {
		return err
	}
	_, err = (*c.conn).Write(packet)
	return err
}

//connect server
func (c *Client) ConnServer() error {
	//check
	if c.conf.Host == "" || c.conf.Port <= 0 {
		return errors.New("host or port is invalid")
	}
	if c.connected {
		return nil
	}

	//format address
	address := fmt.Sprintf("%s:%d", c.conf.Host, c.conf.Port)

	//try connect server
	timeOut := c.conf.ConnTimeOut * time.Second
	conn, err := net.DialTimeout("tcp", address, timeOut)
	if err != nil {
		return err
	}

	//sync conn
	c.conn = &conn
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

//read process
func (c *Client) runReadProcess() {
	var (
		buff = make([]byte, c.conf.ReadBuffSize)
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
			break
		}

		//try read tcp data
		_, err = (*c.conn).Read(buff)
		if err != nil {
			break
		}

		//call cb
		if c.cbForRead != nil {
			c.cbForRead(buff)
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
	_, err := (*c.conn).Write(packBytes)
	return err
}

//inter init
func (c *Client) interInit() {
	//check config
	if c.conf.ReadBuffSize <= 0 {
		c.conf.ReadBuffSize = define.DefaultTcpReadBuffSize
	}
	if c.conf.ConnTimeOut <= 0 {
		c.conf.ConnTimeOut = time.Duration(define.DefaultTcpDialTimeOut)
	}
	if c.conf.WriteTimeOut <= 0 {
		c.conf.WriteTimeOut = define.DefaultTcpWriteTimeOut
	}

	//start delay process
	sf := func() {
		go c.runReadProcess()
	}
	time.AfterFunc(time.Second, sf)
}
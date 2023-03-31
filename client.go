package cree

import (
	"errors"
	"fmt"
	"github.com/andyzhou/cree/define"
	"github.com/andyzhou/cree/face"
	"log"
	"net"
)

/*
 * face for client
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//face info
type Client struct {
	host string
	port int
	readBuffSize int
	conn *net.Conn
	connected bool
	cbForRead func(data []byte) bool
	lazySendChan chan []byte
	closeChan chan bool
}

//construct
func NewClient(
			host string,
			port int,
			lazyChanSize ...int,
		) *Client {
	var (
		lazySendChanSize int
	)

	//setup option parameter
	lazySendChanSize = define.DefaultLazySendChanSize
	if lazyChanSize != nil && len(lazyChanSize) > 0 {
		lazySendChanSize = lazyChanSize[0]
	}

	//self init
	this := &Client{
		host: host,
		port: port,
		readBuffSize: define.DefaultTcpReadBuffSize,
		lazySendChan: make(chan []byte, lazySendChanSize),
		closeChan: make(chan bool, 1),
	}
	go this.runLazySendProcess()
	return this
}

//close connect
func (c *Client) Close() {
	if c.conn == nil {
		return
	}
	(*c.conn).Close()
	c.conn = nil
	c.connected = false
	c.closeChan <- true
	close(c.lazySendChan)
}

//set cb for read data
func (c *Client) SetCBForRead(cb func(data []byte) bool) bool {
	if cb == nil {
		return false
	}
	c.cbForRead = cb
	return true
}

//set read buff size
func (c *Client) SetReadBuffSize(size int) bool {
	if size <= 0 {
		return false
	}
	c.readBuffSize = size
	return true
}

//send packet data
func (c *Client) SendPacket(
					messageId uint32,
					data []byte,
					lazySend ...bool,
				) error {
	var (
		isLazySend bool
	)

	//check
	if messageId < 0 || data == nil {
		return errors.New("invalid parameter")
	}
	if c.conn == nil {
		return errors.New("connect is nil")
	}

	//packet data
	packet := c.packetData(messageId, data)

	//check lazy mode
	if lazySend != nil && len(lazySend) > 0 {
		isLazySend = lazySend[0]
	}

	//try send to server
	if isLazySend {
		select {
		case c.lazySendChan <- packet:
		}
	}else{
		//send direct
		_, err := (*c.conn).Write(packet)
		if err != nil {
			return err
		}
	}
	return nil
}

//connect server
func (c *Client) ConnServer() error {
	//check
	if c.host == "" || c.port <= 0 {
		return errors.New("host or port is invalid")
	}
	if c.connected {
		return nil
	}

	//format address
	address := fmt.Sprintf("%s:%d", c.host, c.port)

	//try connect server
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}

	//sync conn
	c.conn = &conn
	c.connected = true

	//spawn read process
	go c.runReadProcess()

	return nil
}

////////////////
//private func
////////////////

//packet one data
func (c *Client) packetData(messageId uint32, data []byte) []byte {
	message := face.NewMessage()
	message.Id = messageId
	message.SetData(data)
	packet := face.NewPacket()
	byteData, _ := packet.Pack(message)
	return byteData
}

//lazy send process
func (c *Client) runLazySendProcess() {
	var (
		packet []byte
		isOk bool
	)

	//defer
	defer func() {
		if err := recover(); err != nil {
			log.Println("Client:runLazySendProcess panic, err:", err)
		}
		close(c.closeChan)
	}()

	//loop
	for {
		select {
		case packet, isOk = <- c.lazySendChan:
			if isOk {
				(*c.conn).Write(packet)
			}
		case <- c.closeChan:
			return
		}
	}
}

//read process
func (c *Client) runReadProcess() {
	var (
		buff []byte
		err error
	)

	//init buff
	buff = make([]byte, c.readBuffSize)

	//defer
	defer func() {
		if err := recover(); err != nil {
			log.Println("Client:runReadProcess panic, err:", err)
		}
		c.Close()
	}()

	//loop
	for {
		//try read tcp data
		_, err = (*c.conn).Read(buff)
		if err != nil {
			panic(err)
		}

		//call cb
		if c.cbForRead != nil {
			c.cbForRead(buff)
		}
	}
}
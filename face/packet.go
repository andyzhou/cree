package face

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/andyzhou/cree/iface"
)

/*
 * face for data packet
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

 //macro define
 const (
 	PacketHeadSize = 12 //dataLen(4byte) + messageKind(4byte) + messageId(4byte)
 	PacketMaxSize = 4096 //4KB
 )

 //face info
 type Packet struct {
	 littleEndian bool
 }

 //construct
func NewPacket() *Packet {
	//self init
	this := &Packet{
		littleEndian: true,
	}
	return this
}

////////
//api
////////

//set little endian
func (f *Packet) SetLittleEndian(littleEndian bool) {
	f.littleEndian = littleEndian
}

//unpack data, just for message length and id from header
func (f *Packet) UnPack(data []byte) (iface.IMessage, error) {
	var (
		messageKind uint32
		messageId uint32
		messageLen uint32
		err error
	)

	//basic check
	if data == nil || len(data) <= 0 {
		return nil, errors.New("invalid parameter")
	}

	//init data buff
	dataBuff := bytes.NewReader(data)

	//read length
	err = binary.Read(dataBuff, binary.LittleEndian, &messageLen)
	if err != nil {
		return nil, err
	}

	//read message kind
	err = binary.Read(dataBuff, binary.LittleEndian, &messageKind)
	if err != nil {
		return nil, err
	}

	//read message id
	err = binary.Read(dataBuff, binary.LittleEndian, &messageId)
	if err != nil {
		return nil, err
	}

	//read data
	if messageLen > PacketMaxSize {
		tips := fmt.Sprintf("too large message data received, message length:%d", messageLen)
		return nil, errors.New(tips)
	}

	//init message data
	message := NewMessage()
	message.SetKind(messageKind)
	message.SetId(messageId)
	message.SetLen(messageLen)

	return message, nil
}

//pack data
func (f *Packet) Pack(message iface.IMessage) ([]byte, error) {
	var (
		err error
	)

	//basic check
	if message == nil {
		return nil, errors.New("invalid parameter")
	}

	//init data buff
	dataBuff := bytes.NewBuffer(nil)

	//write length
	err = binary.Write(dataBuff, binary.LittleEndian, message.GetLen())
	if err != nil {
		return nil, err
	}

	//write kind
	err = binary.Write(dataBuff, binary.LittleEndian, message.GetKind())
	if err != nil {
		return nil, err
	}

	//write message id
	err = binary.Write(dataBuff, binary.LittleEndian, message.GetId())
	if err != nil {
		return nil, err
	}

	//write data
	err = binary.Write(dataBuff, binary.LittleEndian, message.GetData())
	if err != nil {
		return nil, err
	}

	return dataBuff.Bytes(), nil
}

//get length
func (f *Packet) GetHeadLen() uint32 {
	return PacketHeadSize
}



package face

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/andyzhou/cree/define"
	"github.com/andyzhou/cree/iface"
)

/*
 * face for data packet
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//inter macro define
//DO NOT CHANGE THIS!!!
const (
	PacketHeadSize = 12 //dataLen(4byte) + messageKind(4byte) + messageId(4byte)
)

//face info
type Packet struct {
	maxPackSize  int
	littleEndian bool
	byteOrder    binary.ByteOrder
}

 //construct
func NewPacket() *Packet {
	//self init
	this := &Packet{
		maxPackSize: define.PacketMaxSize,
		littleEndian: true,
		byteOrder: binary.LittleEndian,
	}
	return this
}

//set max pack size
func (f *Packet) SetMaxPackSize(size int) {
	if size <= 0 {
		return
	}
	f.maxPackSize = size
}

//set big or little endian
func (f *Packet) SetLittleEndian(littleEndian bool) {
	f.littleEndian = littleEndian
	if littleEndian {
		f.byteOrder = binary.LittleEndian
	}else{
		f.byteOrder = binary.BigEndian
	}
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
	err = binary.Read(dataBuff, f.byteOrder, &messageLen)
	if err != nil {
		return nil, err
	}

	//read message kind
	err = binary.Read(dataBuff, f.byteOrder, &messageKind)
	if err != nil {
		return nil, err
	}

	//read message id
	err = binary.Read(dataBuff, f.byteOrder, &messageId)
	if err != nil {
		return nil, err
	}

	//read data
	if messageLen > uint32(f.maxPackSize) {
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
	err = binary.Write(dataBuff, f.byteOrder, message.GetLen())
	if err != nil {
		return nil, err
	}

	//write kind
	err = binary.Write(dataBuff, f.byteOrder, message.GetKind())
	if err != nil {
		return nil, err
	}

	//write message id
	err = binary.Write(dataBuff, f.byteOrder, message.GetId())
	if err != nil {
		return nil, err
	}

	//write data
	err = binary.Write(dataBuff, f.byteOrder, message.GetData())
	if err != nil {
		return nil, err
	}

	return dataBuff.Bytes(), nil
}

//get length
func (f *Packet) GetHeadLen() uint32 {
	return PacketHeadSize
}



package iface

//interface of packet
type IPacket interface {
	UnPack(data []byte) (IMessage, error)
	Pack(message IMessage) ([]byte, error)
	GetHeadLen() uint32
	SetLittleEndian(littleEndian bool)
	SetMaxPackSize(size int)
}

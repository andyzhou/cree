package face

/*
 * face for message data
 */

//face info
type Message struct {
	Len uint32
	Id uint32
	Data []byte
}

//construct
func NewMessage() *Message {
	//self init
	this := &Message{}
	return this
}

///////
//api
///////

//get relate data
func (f *Message) GetLen() uint32 {
	return f.Len
}

func (f *Message) GetId() uint32 {
	return f.Id
}

func (f *Message) GetData() []byte {
	return f.Data
}

//set relate data
func (f *Message) SetId(id uint32) {
	f.Id = id
}

func (f *Message) SetLen(len uint32) {
	f.Len = len
}

func (f *Message) SetData(data []byte) {
	f.Data = data
	f.Len = uint32(len(data))
}
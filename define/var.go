package define

type (
	//send message request
	SendMsgReq struct {
		MsgId    uint32
		Data     []byte
		ConnIds  []int64
		Tags     []string
		Property map[string]interface{}
	}
)

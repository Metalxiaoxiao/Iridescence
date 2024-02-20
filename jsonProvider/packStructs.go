package jsonprovider

// StandardJSONPack 根数据包结构体，仅用于websocket
type StandardJSONPack struct {
	Command string `json:"command"`
}

type LoginPackage struct {
	Userid   int    `json:"userId"`
	Password string `json:"password"`
}
type LoginPackageRes struct {
	State   bool
	Message string
}
type SignUpPackage struct {
	Userid   int    `json:"userid"`
	Password string `json:"password"`
}
type SendMessageRequestPack struct {
	TargetID         int    `json:"targetId"`    //消息接收人
	RequestID        int    `json:"requestId"`   //request ID由客户端生成
	MessageBody      string `json:"messageBody"` //消息体
	RequestTimeStamp int    `json:"time"`        //判断请求是否合法，是否超时
}

// SendMessageRequestPackRes 实现ACK机制
type SendMessageRequestPackRes struct {
	RequestID int `json:"requestId"` //返回requestID，用于ACK机制
	MessageID int `json:"messageId"` //返回递增的数据库主键，作为MessageID,用户可以用messageID进行后续的撤回，引用等操作
	TimeStamp int `json:"time"`
	State     int `json:"state"` //是否成功
}

const (
	UserRefused = iota
	ServerSendError
	UserIsNotOnline
	UserReceived
)

type SendMessageToTargetPack struct {
	SenderID    int    `json:"senderId"`
	MessageID   int    `json:"messageId"`
	MessageBody string `json:"messageBody"`
	TimeStamp   int    `json:"time"`
}
type SendMessagePackResFromUser struct {
	MessageID int  `json:"messageId"`
	State     bool `json:"state"`
}

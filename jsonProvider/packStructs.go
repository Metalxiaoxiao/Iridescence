package jsonprovider

// 数据包结构体
type StandardJSONPack struct {
	Command string `json:"command"`
}
type LoginPackage struct {
	Userid   int    `json:"userid"`
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
	TargetID         int    `json:"targetid"`
	RequestID        int    `json:"requestid"`
	MessageBody      string `json:"messagebody"`
	RequestTimeStamp int    `json:"time"` //判断请求是否合法，是否超时
}
type SendMessageRequestPackRes struct {
	RequestID int  `json:"requestid"`
	MessageID int  `json:"messageid"`
	TimeStamp int  `json:"time"`
	State     bool `json:"state"`
}
type SendMessagePack struct {
	SenderID    int    `json:"senderid"`
	MessageID   int    `json:"messageid"`
	MessageBody string `json:"messagebody"`
	TimeStamp   int    `json:"time"`
}
type SendMessagePackResFromUser struct {
	MessageID int  `json:"messageid"`
	State     bool `json:"state"`
}

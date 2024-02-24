package jsonprovider

import "encoding/json"

// StandardJSONPack 根数据包结构体，仅用于websocket
type StandardJSONPack struct {
	Command string `json:"command"`
}

type LoginRequest struct {
	Userid   int    `json:"userId"`
	Password string `json:"password"`
}
type LoginResponse struct {
	State   bool
	Message string
}
type SignUpRequest struct {
	UserName int    `json:"userName"`
	Password string `json:"password"`
}
type SendMessageRequest struct {
	TargetID         int    `json:"targetId"`    //消息接收人
	RequestID        int    `json:"requestId"`   //request ID由客户端生成
	MessageBody      string `json:"messageBody"` //消息体
	RequestTimeStamp int    `json:"time"`        //判断请求是否合法，是否超时
}

// SendMessageResponse 实现ACK机制
type SendMessageResponse struct {
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
type SendMessagePackResponseFromUser struct {
	MessageID int  `json:"messageId"`
	State     bool `json:"state"`
}

type AddFriendRequest struct {
	FriendID int `json:"friendId"`
}

type DeleteFriendRequest struct {
	FriendID int `json:"friendId"`
}

type GetUserDataRequest struct {
	UserID int `json:"userId"`
}

type GetUserDataResponse struct {
	UserID         int             `json:"userId"`
	UserName       string          `json:"userName"`
	UserAvatar     string          `json:"userAvatar"`
	UserNote       string          `json:"userNote"`
	UserPermission uint            `json:"userPermission"`
	UserFriendList json.RawMessage `json:"userFriendList"`
}

type ChangeAvatarRequest struct {
	NewAvatar string `json:"newAvatar"`
}

type ChangeAvatarResponse struct {
	UserID    int    `json:"userId"`
	NewAvatar string `json:"newAvatar"`
	Success   bool   `json:"success"`
}

type GetMessagesWithUserRequest struct {
	OtherUserID int `json:"otherUserId"`
	StartTime   int `json:"startTime"`
	EndTime     int `json:"endTime"`
}

type GetMessagesWithUserResponse struct {
	UserID   int       `json:"userId"`
	Messages []Message `json:"messages"`
}

type Message struct {
	MessageID   int    `json:"messageId"`
	SenderID    int    `json:"senderId"`
	ReceiverID  int    `json:"receiverId"`
	Time        int    `json:"time"`
	MessageBody string `json:"messageBody"`
	MessageType int    `json:"messageType"`
}

type CreateGroupRequest struct {
	GroupName         string `json:"groupName"`
	GroupExplaination string `json:"groupExplaination"`
}

type CreateGroupResponse struct {
	GroupID int64 `json:"groupId"`
	Success bool  `json:"success"`
}
type BreakGroupRequest struct {
	GroupID int64 `json:"groupId"`
}

type BreakGroupResponse struct {
	GroupID int64 `json:"groupId"`
	Success bool  `json:"success"`
}
type SendGroupMessageRequest struct {
	GroupID     int64  `json:"groupId"`
	MessageBody string `json:"messageBody"`
	RequestID   int    `json:"requestId"`
}

type SendGroupMessageResponse struct {
	RequestID int `json:"requestId"`
	MessageID int `json:"messageId"`
	TimeStamp int `json:"timeStamp"`
	State     int `json:"state"`
}

type SendMessageToGroupPack struct {
	SenderID    int    `json:"senderId"`
	MessageID   int    `json:"messageId"`
	MessageBody string `json:"messageBody"`
	TimeStamp   int    `json:"timeStamp"`
}

type AddFriendResponse struct {
	UserID   int  `json:"userId"`
	FriendID int  `json:"friendId"`
	Success  bool `json:"success"`
}

type DeleteFriendResponse struct {
	UserID   int  `json:"userId"`
	FriendID int  `json:"friendId"`
	Success  bool `json:"success"`
}
type GetOfflineMessagesResponse struct {
	UserID   int       `json:"userId"`
	Messages []Message `json:"messages"`
}

package jsonprovider

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"time"
)

type User struct {
	UserID         int
	TokenExpiry    time.Time
	Conn           *websocket.Conn
	UserName       string          `json:"userName"`
	UserAvatar     string          `json:"userAvatar"`
	UserNote       string          `json:"userNote"`
	UserPermission uint            `json:"userPermission"`
	UserFriendList json.RawMessage `json:"userFriendList"`
}
type Friend struct {
	UserID  int
	AddTime int
}
type FriendList []Friend

type UserSettings struct {
}
type Group struct {
	GroupName string
	GroupID   int
}
type GroupList []Group

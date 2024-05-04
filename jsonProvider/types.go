package jsonprovider

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
)

type User struct {
	UserId         int             `json:"userId"`
	TokenExpiry    time.Time       `json:"-"`
	Conn           *websocket.Conn `json:"-"`
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

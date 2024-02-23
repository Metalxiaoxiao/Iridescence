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
	Username       string
	UserName       string          `json:"userName"`
	UserAvatar     string          `json:"userAvatar"`
	UserNote       string          `json:"userNote"`
	UserPermission uint            `json:"userPermission"`
	UserFriendList json.RawMessage `json:"userFriendList"`
}

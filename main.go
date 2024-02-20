package main

import (
	"config"
	"database/sql"
	"dbUtils"
	"encoding/json"
	fileserver "filesystem"
	"fmt"
	"hashUtils"
	"httpService"
	jsonprovider "jsonProvider"
	"logger"
	"net/http"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/gorilla/websocket"
)

var (
	clients     = make(map[int]*User) // 保存用户ID与用户结构体的映射关系
	clientsLock sync.Mutex            // 用于保护映射关系的互斥锁

	////用于ACK的消息池
	//processingStateMessages     = make(map[int64]*Message)
	//processingStateMessagesLock sync.Mutex
)

const (
	_VERSION = "0.0.1"
)

// User 用户结构体
type User struct {
	UserID         int
	Conn           *websocket.Conn
	Username       string
	UserName       string          `json:"userName"`
	UserAvatar     string          `json:"userAvatar"`
	UserNote       string          `json:"userNote"`
	UserPermission uint            `json:"userPermission"`
	UserFriendList json.RawMessage `json:"userFriendList"`
}

//Message 消息结构体,用于临时消息池
//type Message struct {
//	tempID      int
//	id          int
//	messageBody string
//}

var (
	confData config.Config
	db       *sql.DB
)

const (
	UserMessage = iota
	SystemMessage
)

func main() {

	confData, err := config.LoadConfig("config.json")
	if err != nil {
		logger.Error(err)
	}
	logger.Info("日志级别被设置为", confData.LogLevel)
	pink := "\033[96m" // 天蓝色 ANSI 转义序列
	reset := "\033[0m" // 重置颜色 ANSI 转义序列
	fmt.Println(``)
	fmt.Println(pink + "  _____      _     _                                  ")
	fmt.Println("  \\_   \\_ __(_) __| | ___ ___  ___ ___ _ __   ___ ___")
	fmt.Println("   / /\\/ '__| |/ _` |/ _ \\ __|/ __/ _ \\ '_ \\ / __/ _ \\")
	fmt.Println("/\\/ /_ | |  | | (_| |  __\\__ \\ (__  __/ | | | (__  __/")
	fmt.Println("\\____/ |_|  |_|\\__,_|\\___|___/\\___\\___|_| |_|\\___\\___|")
	fmt.Println(reset)
	fmt.Println("     - A Stable and Highly Available Chat Server -")
	fmt.Println("VERSION:", _VERSION, " , Built on 2023/6/26")
	logger.SetLogLevel(logger.LogLevel(confData.LogLevel))

	var (
		_PROT = confData.Port
	)

	logger.Info("Server running at port:", _PROT)
	logger.Info("Trying to connect to MYSQL...")

	if confData.DataBaseSettings.Account == "default_account" {
		logger.Error("配置文件中数据库信息未修改，请修改后再启动程序")
		time.Sleep(time.Second * 5)
		return
	}

	dbUtils.DbInit(confData)
	hashUtils.LoadConfig(confData)
	httpService.LoadConfig(confData)

	db = dbUtils.GetDBPtr()

	logger.Info("服务器启动成功！")
	http.HandleFunc(confData.WebSocketServiceRote, handleWebSocket)
	http.HandleFunc(confData.RegisterServiceRote, httpService.HandleRegister)
	http.HandleFunc(confData.LoginServiceRote, httpService.HandleLogin)
	http.HandleFunc(confData.UploadServiceRote, fileserver.HandleFileUpload)
	http.HandleFunc(confData.DownloadServiceRote, fileserver.HandleFileDownload)
	logger.Error(http.ListenAndServe(":"+_PROT, nil))
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {

	// 完成WebSocket握手
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("WebSocket upgrade failed:", err)
		return
	}
	Logined := false
	var userID int
	// 处理WebSocket消息
	for !Logined {
		// 用户登录过程
		// 在此处获取用户ID，并保存到映射关系中
		var res jsonprovider.LoginResponse
		var p jsonprovider.LoginRequest
		err = conn.ReadJSON(&p)
		logger.Debug(p)
		if err != nil {
			logger.Error("用户登录时读取消息失败", err)
		}
		userID = p.Userid

		passwordHash, passwordSalt, err := dbUtils.GetDBPasswordHash(userID)
		if err != nil {
			logger.Error("读取数据库密码哈希值失败", err)
		}
		logger.Debug("登录时读取盐:", passwordSalt)
		tryingPasswordHash := hashUtils.HashPassword(p.Password, passwordSalt)
		logger.Debug("尝试哈希", tryingPasswordHash, "实际哈希", passwordHash)
		if tryingPasswordHash == passwordHash {
			// 从数据库中获取用户信息
			var username, userAvatar, userNote string
			var userPermission uint
			var userFriendList json.RawMessage
			err := db.QueryRow("SELECT userName, userAvatar, userNote, userPermission, userFriendList FROM userdatatable WHERE userID = ?", userID).Scan(&username, &userAvatar, &userNote, &userPermission, &userFriendList)
			if err != nil {
				logger.Error("获取用户数据失败:", err)
				return
			}

			// 创建新的User结构体
			user := &User{
				UserID:         userID,
				Conn:           conn,
				Username:       username,
				UserName:       username,
				UserAvatar:     userAvatar,
				UserNote:       userNote,
				UserPermission: userPermission,
				UserFriendList: userFriendList,
			}

			// 保存到clients map中
			clientsLock.Lock()
			clients[userID] = user
			clientsLock.Unlock()

			res = jsonprovider.LoginResponse{
				State:   true,
				Message: "登录成功",
			}
			logger.Debug("用户", userID, "登录成功")
			Logined = true
		} else {
			res = jsonprovider.LoginResponse{
				State:   false,
				Message: "登录失败",
			}
		}
		message := jsonprovider.StringifyJSON(res)
		err = conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			logger.Error("Failed to send message:", err)
			// 处理发送消息失败的情况
		}
	}

	//消息处理主循环
	for {
		// 读取消息
		_, message, err := conn.ReadMessage()
		if err != nil {
			logger.Debug("读取消息失败，可能是用户断开连接:", err)
			break
		}

		// 在这里处理消息，用保存的映射关系来识别和处理特定用户的消息
		logger.Debug("Received message from user", userID, ":", string(message), "\n")
		var pre jsonprovider.StandardJSONPack
		jsonprovider.ParseJSON(message, &pre)
		switch pre.Command {
		case "sendUserMessage":
			var state int
			//获取基本信息
			var receivedPack jsonprovider.SendMessageRequest
			jsonprovider.ParseJSON(message, &receivedPack)
			recipientID := receivedPack.TargetID
			messageContent := receivedPack.MessageBody
			requestMessageID := receivedPack.RequestID
			timeStamp := int(time.Now().UnixNano())
			//保存到数据库，获取消息ID
			var messageID int
			messageID, err = dbUtils.SaveOfflineMessageToDB(userID, recipientID, messageContent, UserMessage)
			if err != nil {
				logger.Error("用户", recipientID, "发送信息时数据库插入失败")
				break
			}
			//构造发送数据包
			sendingPack := &jsonprovider.SendMessageToTargetPack{
				SenderID:    userID,
				MessageID:   messageID,
				MessageBody: messageContent,
				TimeStamp:   timeStamp,
			}
			// 向指定用户发送消息
			isSent, msgerr := sendMessageToUser(recipientID, []byte(jsonprovider.StringifyJSON(sendingPack)))
			if !isSent {
				if msgerr == nil {
					logger.Info("用户", recipientID, "不在线，已保存到离线消息")
					state = jsonprovider.UserIsNotOnline
				} else {
					state = jsonprovider.ServerSendError
				}
			} else {
				state = jsonprovider.UserReceived
			}
			//回发ACK包
			ACKPack := &jsonprovider.SendMessageResponse{
				RequestID: requestMessageID,
				MessageID: messageID,
				TimeStamp: timeStamp,
				State:     state,
			}
			_, err := sendMessageToUser(userID, []byte(jsonprovider.StringifyJSON(ACKPack)))
			if err != nil {
				logger.Debug("ACK回发错误", err)
				break
			}
		case "sendGroupMessage":
			var req jsonprovider.SendGroupMessageRequest
			jsonprovider.ParseJSON(message, &req)

			// 保存消息到数据库
			timeStamp := int(time.Now().UnixNano())
			var messageID int
			messageID, err = dbUtils.SaveOfflineGroupMessageToDB(userID, int(req.GroupID), req.MessageBody, UserMessage)
			if err != nil {
				logger.Error("用户发送群消息时数据库插入失败")
				break
			}

			// 构造发送数据包
			sendingPack := &jsonprovider.SendMessageToGroupPack{
				SenderID:    userID,
				MessageID:   messageID,
				MessageBody: req.MessageBody,
				TimeStamp:   timeStamp,
			}

			// 获取群成员
			var groupMembers []int
			err = db.QueryRow("SELECT groupMembers FROM groupdatatable WHERE groupID = ?", req.GroupID).Scan(&groupMembers)
			if err != nil {
				logger.Error("Failed to get group members:", err)
				return
			}

			// 向所有群成员发送消息
			for _, memberID := range groupMembers {
				_, err := sendMessageToUser(memberID, []byte(jsonprovider.StringifyJSON(sendingPack)))
				if err != nil {
					logger.Debug("群消息发送错误", err)
					break
				}
			}

			//回发ACK包
			ACKPack := &jsonprovider.SendGroupMessageResponse{
				RequestID: req.RequestID,
				MessageID: messageID,
				TimeStamp: timeStamp,
				State:     jsonprovider.UserReceived,
			}
			_, err = sendMessageToUser(userID, []byte(jsonprovider.StringifyJSON(ACKPack)))
			if err != nil {
				logger.Debug("群消息ACK回发错误", err)
				break
			}
		case "addFriend":
			var req jsonprovider.AddFriendRequest
			jsonprovider.ParseJSON(message, &req)

			// 获取用户的朋友列表
			friendList := clients[userID].UserFriendList
			var friends []int
			err := json.Unmarshal(friendList, &friends)
			if err != nil {
				break
			}

			// 添加新朋友
			friends = append(friends, req.FriendID)

			// 更新朋友列表
			newFriendList, _ := json.Marshal(friends)
			clients[userID].UserFriendList = newFriendList

			// 更新数据库
			_, err = db.Exec("UPDATE userdatatable SET userFriendList = ? WHERE userID = ?", newFriendList, userID)
			if err != nil {
				logger.Error("更新好友数据失败:", err)
			}
		case "deleteFriend":
			var req jsonprovider.DeleteFriendRequest
			jsonprovider.ParseJSON(message, &req)

			// 获取用户的朋友列表
			friendList := clients[userID].UserFriendList
			var friends []int
			err := json.Unmarshal(friendList, &friends)
			if err != nil {
				break
			}

			// 删除朋友
			for i, friend := range friends {
				if friend == req.FriendID {
					friends = append(friends[:i], friends[i+1:]...)
					break
				}
			}

			// 更新朋友列表
			newFriendList, _ := json.Marshal(friends)
			clients[userID].UserFriendList = newFriendList

			// 更新数据库
			_, err = db.Exec("UPDATE userdatatable SET userFriendList = ? WHERE userID = ?", newFriendList, userID)
			if err != nil {
				logger.Error("Failed to update friend list:", err)
			}
		case "changeFriendSettings":

		case "createGroup":
			var req jsonprovider.CreateGroupRequest
			jsonprovider.ParseJSON(message, &req)

			// 在数据库中创建新的群聊
			res, err := db.Exec("INSERT INTO groupdatatable (groupName, groupExplaination, groupMaster) VALUES (?, ?, ?)", req.GroupName, req.GroupExplaination, userID)
			if err != nil {
				logger.Error("Failed to create group:", err)
				return
			}

			// 获取新群聊的ID
			groupID, err := res.LastInsertId()
			if err != nil {
				logger.Error("Failed to get group ID:", err)
				return
			}

			// 创建新的群聊成员列表
			groupMembers := []int{userID}

			// 更新群聊的成员列表
			groupMembersJSON, _ := json.Marshal(groupMembers)
			_, err = db.Exec("UPDATE groupdatatable SET groupMembers = ? WHERE groupID = ?", groupMembersJSON, groupID)
			if err != nil {
				logger.Error("Failed to update group members:", err)
			}

			// 创建响应
			responsePack := jsonprovider.CreateGroupResponse{
				GroupID: groupID,
				Success: err == nil,
			}

			// 发送响应
			message := jsonprovider.StringifyJSON(responsePack)
			_, err = sendMessageToUser(userID, []byte(message))
			if err != nil {
				logger.Error("Failed to send group creation response:", err)
			}
		case "breakGroup":
			var req jsonprovider.BreakGroupRequest
			jsonprovider.ParseJSON(message, &req)

			// 在数据库中删除群聊
			_, err := db.Exec("DELETE FROM groupdatatable WHERE groupID = ? AND groupMaster = ?", req.GroupID, userID)
			if err != nil {
				logger.Error("Failed to break group:", err)
				return
			}

			// 创建响应
			res := jsonprovider.BreakGroupResponse{
				GroupID: req.GroupID,
				Success: err == nil,
			}

			// 发送响应
			message := jsonprovider.StringifyJSON(res)
			_, err = sendMessageToUser(userID, []byte(message))
			if err != nil {
				logger.Error("Failed to send group break response:", err)
			}

		case "changeGroupSettings":

		case "getUserData":
			var req jsonprovider.GetUserDataRequest
			jsonprovider.ParseJSON(message, &req)

			// 从数据库中获取用户数据
			var username, userAvatar, userNote string
			var userPermission uint
			var userFriendList json.RawMessage
			err := db.QueryRow("SELECT userName, userAvatar, userNote, userPermission, userFriendList FROM userdatatable WHERE userID = ?", req.UserID).Scan(&username, &userAvatar, &userNote, &userPermission, &userFriendList)
			if err != nil {
				logger.Error("Failed to get user data:", err)
				return
			}

			// 创建响应
			res := jsonprovider.GetUserDataResponse{
				UserID:         req.UserID,
				UserName:       username,
				UserAvatar:     userAvatar,
				UserNote:       userNote,
				UserPermission: userPermission,
				UserFriendList: userFriendList,
			}

			// 发送响应
			message := jsonprovider.StringifyJSON(res)
			_, err = sendMessageToUser(userID, []byte(message))
			if err != nil {
				logger.Error("Failed to send user data:", err)
			}
		case "messageEvent":

		case "userStateEvent":

		case "getUnreceivedMessage":

		case "getMessagesWithUser":
			var req jsonprovider.GetMessagesWithUserRequest
			jsonprovider.ParseJSON(message, &req)

			// 从数据库中查询聊天记录
			rows, err := db.Query("SELECT messageID, senderID, receiverID, time, messageBody, messageType FROM offlinemessages WHERE ((senderID = ? AND receiverID = ?) OR (senderID = ? AND receiverID = ?)) AND time BETWEEN ? AND ?", userID, req.OtherUserID, req.OtherUserID, userID, req.StartTime, req.EndTime)
			if err != nil {
				logger.Error("Failed to get messages:", err)
				return
			}

			// 读取聊天记录
			var messages []jsonprovider.Message
			for rows.Next() {
				var message jsonprovider.Message
				err := rows.Scan(&message.MessageID, &message.SenderID, &message.ReceiverID, &message.Time, &message.MessageBody, &message.MessageType)
				if err != nil {
					logger.Error("Failed to read message:", err)
					return
				}
				messages = append(messages, message)
			}
			err = rows.Close()
			if err != nil {

			}
			// 创建响应
			res := jsonprovider.GetMessagesWithUserResponse{
				UserID:   userID,
				Messages: messages,
			}

			// 发送响应
			message := jsonprovider.StringifyJSON(res)
			_, err = sendMessageToUser(userID, []byte(message))
			if err != nil {
				logger.Error("Failed to send message history:", err)
			}

		case "changeSettings":

		case "postOpreation":

		case "changeAvatar":
			var req jsonprovider.ChangeAvatarRequest
			jsonprovider.ParseJSON(message, &req)

			// 更新用户结构体
			clients[userID].UserAvatar = req.NewAvatar

			// 更新数据库
			_, err := db.Exec("UPDATE userdatatable SET userAvatar = ? WHERE userID = ?", req.NewAvatar, userID)
			if err != nil {
				logger.Error("Failed to update avatar:", err)
			}

			// 创建响应
			res := jsonprovider.ChangeAvatarResponse{
				UserID:    userID,
				NewAvatar: req.NewAvatar,
				Success:   err == nil,
			}

			// 发送响应
			message := jsonprovider.StringifyJSON(res)
			_, err = sendMessageToUser(userID, []byte(message))
			if err != nil {
				logger.Error("Failed to send avatar change response:", err)
			}
		}

	}

	// 用户断开连接
	// 在此处删除映射关系
	if Logined {
		clientsLock.Lock()
		delete(clients, userID)
		clientsLock.Unlock()
	}

}

func broadcastMessage(message []byte) {
	clientsLock.Lock()
	defer clientsLock.Unlock()

	for _, client := range clients {
		err := client.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			logger.Error("Failed to send message:", err)
			// 处理发送消息失败的情况
		}
	}
}

func sendMessageToUser(userID int, message []byte) (bool, error) {
	clientsLock.Lock()
	defer clientsLock.Unlock()

	client, ok := clients[userID]
	if !ok {
		logger.Warn("用户不在线:", userID)
		return false, nil
	}

	err := client.Conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		logger.Error("消息发送失败:", err)
		// 处理发送消息失败的情况
		return false, err
	}

	return true, nil
}

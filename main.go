package main

import (
	"config"
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

var confData config.Config

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

	logger.Info("服务器启动成功！")
	http.HandleFunc(confData.WebSocketServiceRote, handleWebSocket)
	http.HandleFunc(confData.RegisterServiceRote, httpService.HandleRegister)
	http.HandleFunc(confData.UploadServiceRote, fileserver.HandleFileUpload)
	http.HandleFunc(confData.DownloadServiceRote, fileserver.HandleFileDownload)
	logger.Error(http.ListenAndServe(":"+_PROT, nil))
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {

	// 完成WebSocket握手
	upgrader := websocket.Upgrader{}
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
		var res jsonprovider.LoginPackageRes
		var p jsonprovider.LoginPackage
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
			clientsLock.Lock()

			clients[userID] = &User{
				UserID: p.Userid,
				Conn:   conn,
			}
			clientsLock.Unlock()
			res = jsonprovider.LoginPackageRes{
				State:   true,
				Message: "登录成功",
			}
			logger.Debug("用户", userID, "登录成功")
			Logined = true

		} else {
			res = jsonprovider.LoginPackageRes{
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
			/*发送消息确保不丢包：
			1.客户端发送数据包到Server，该数据包仅包含一个临时ID（可以是客户端生成的时间戳），接收者和基本消息内容。
			2.服务端收到发送者数据包后插入Session数据库，并为消息分配一个ID（服务端的时间戳）
			3.服务端返回发送成功数据包（包含一个时间戳和一个消息ID)
			4.服务端向接收者发送数据包（包含时间戳和消息ID)，并等待接收者发回接收成功数据包；
			如果超时未收到接收成功包，那么尝试重新发送10次，10次还未发送成功，向发送者返回发送失败数据包。
			5.如果接收者不在线，存入Session会话储存数据库
			*/
			// var decodedPack jsonprovider.SendMessageRequestPack
			// jsonprovider.ParseJSON(message, &decodedPack)
			// insertQuery := "INSERT INTO chatdata (conversation_id, user_id, recipient_id, content, timestamp) VALUES (?, ?, ?, ?, ?)"
			// timestamp := time.Now()
			// recipientID := decodedPack.TargetID
			// messageContent := decodedPack.MessageBody
			// useDB(_SessionDBName)
			// // _, err = db.Exec(, insertQuery, userID, recipientID, messageContent, timestamp)
			// if err != nil {
			// 	logger.Error("保存用户漫游消息时出现错误", err)
			// }

		case "sendGroupMessage":

		case "addFriend":

		case "deleteFriend":

		case "changeFriendSettings":

		case "createGroup":

		case "breakGroup":

		case "changeGroupSettings":

		case "getUserData":

		case "messageEvent":

		case "userStateEvent":

		case "getUnreceivedMessage":

		case "getSessionMessage":

		case "changeSettings":

		case "postOpreation":

		case "changeAvatar":

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

func sendMessageToUser(userID int, message []byte) {

	clientsLock.Lock()
	defer clientsLock.Unlock()

	client, ok := clients[userID]
	if !ok {
		logger.Error("User not found:", userID)
		// 处理用户不存在的情况
		return
	}

	err := client.Conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		logger.Error("Failed to send message:", err)
		// 处理发送消息失败的情况
	}
}

package main

import (
	"config"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	fileserver "filesystem"
	"fmt"
	jsonprovider "jsonProvider"
	"logger"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/gorilla/websocket"
)

var (
	clients     = make(map[int]*User) // 保存用户ID与用户结构体的映射关系
	clientsLock sync.Mutex            // 用于保护映射关系的互斥锁
)
var (
	_BasicChatDBName = "basic_chat_base"
	_SessionDBName   = "session_data_base"
)
var db *sql.DB

const (
	_VERSION   = "0.0.1"
	saltLength = 8 // 密码加盐的长度，可以根据需要进行调整
)

// 用户结构体
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

func useDB(DBname string) {
	_, err := db.Exec("USE " + DBname)
	if err != nil {
		logger.Error("操作 ", DBname, " 数据库时出现错误:", err)
	}
}

func checkTableExistence(DBname string, tableName string) int {
	useDB("information_schema")
	query := "SELECT COUNT(*) FROM tables WHERE table_schema = ? AND table_name = ?"
	var tablecount int
	err := db.QueryRow(query, DBname, tableName).Scan(&tablecount)
	if err != nil {
		logger.Error("Failed to check table existence:", err)
	}
	return tablecount
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
	db, err = sql.Open("mysql", confData.DataBaseSettings.Account+":"+confData.DataBaseSettings.Password+"@tcp("+confData.DataBaseSettings.Address+")/")
	if err != nil {
		logger.Error(err)
	} else {
		logger.Info("Connected to MYSQL successfully")
	}
	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS " + _BasicChatDBName)
	if err != nil {
		logger.Error("创建 ", _BasicChatDBName, " 数据库时出现错误:", err)
	}

	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS " + _BasicChatDBName)
	if err != nil {
		logger.Error("创建 ", _SessionDBName, " 数据库时出现错误:", err)
	}

	//检查表的存在性并建表
	{

		// 如果表不存在，则创建表
		if checkTableExistence(_BasicChatDBName, "userdatatable") == 0 {
			useDB(_BasicChatDBName)
			logger.Warn("找不到用户数据表，自动创建")
			createTable := `
		CREATE TABLE userdatatable (
			userID int unsigned NOT NULL AUTO_INCREMENT,
			userName varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
			userAvatar text,
			userNote varchar(255) DEFAULT NULL,
			userPermission int unsigned DEFAULT 4,
			userFriendList json DEFAULT NULL,
			userGroupList json DEFAULT NULL,
			userHomePageData json DEFAULT NULL,
			userSettings json DEFAULT NULL,
			userPasswordHashValue text,
			passwordSalt BINARY(` + strconv.Itoa(saltLength) + `),
			PRIMARY KEY (userID)
		  ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
		`
			_, err = db.Exec(createTable)
			if err != nil {
				logger.Error("Failed to create table:", err)
			}
		}
		if checkTableExistence(_BasicChatDBName, "groupdatatable") == 0 {
			useDB(_BasicChatDBName)
			logger.Warn("找不到群聊数据表，自动创建")
			createTable := `
			CREATE TABLE groupdatatable (
				groupID int NOT NULL AUTO_INCREMENT,
				groupName varchar(255) NOT NULL,
				groupAvatar varchar(255) DEFAULT NULL,
				groupExplaination text NOT NULL,
				groupMaster int DEFAULT NULL,
				groupMembers json DEFAULT NULL,
				groupSettings json DEFAULT NULL,
				PRIMARY KEY (groupID)
			  ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
		`
			_, err = db.Exec(createTable)
			if err != nil {
				logger.Error("Failed to create table:", err)
			}
		}
		if checkTableExistence(_SessionDBName, "offlinemessages") == 0 {
			useDB(_SessionDBName)
			logger.Warn("找不到离线消息数据表，自动创建")
			createTable := `CREATE TABLE offlinemessages (
				receiverID int unsigned NOT NULL,
				messageID int unsigned DEFAULT NULL,
				time datetime DEFAULT NULL,
				senderID int unsigned DEFAULT NULL,
				messageData json DEFAULT NULL,
				messageType smallint unsigned DEFAULT NULL,
				PRIMARY KEY (receiverID) USING BTREE
			  ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;`
			_, err = db.Exec(createTable)
			if err != nil {
				logger.Error("Failed to create table:", err)
			}
		}
		if checkTableExistence(_SessionDBName, "chatdata") == 0 {
			useDB(_SessionDBName)
			logger.Warn("找不到聊天漫游数据表，自动创建")
			createTable := `CREATE TABLE chatdata (
				message_id int unsigned NOT NULL AUTO_INCREMENT,
				conversation_id int NOT NULL,
				user_id int DEFAULT NULL,
				recipient_id int DEFAULT NULL,
				content text,
				timestamp datetime DEFAULT NULL,
				PRIMARY KEY (message_id),
				KEY idx_conversation_id (conversation_id),
				KEY idx_user_id (user_id),
				KEY idx_recipient_id (recipient_id),
				KEY idx_timestamp (timestamp)
			  ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;`
			_, err = db.Exec(createTable)
			if err != nil {
				logger.Error("Failed to create table:", err)
			}
		}
	}
	logger.Info("服务器启动成功！")
	http.HandleFunc(confData.WebSocketServiceRote, handleWebSocket)
	http.HandleFunc(confData.RegisterServiceRote, handleRegister)
	http.HandleFunc(confData.UploadServiceRote, fileserver.HandleFileUpload)
	http.HandleFunc(confData.DownloadServiceRote, fileserver.HandleFileDownload)
	logger.Error(http.ListenAndServe(":"+_PROT, nil))
}

// 注册逻辑
func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "不允许GET请求，请使用POST重新请求")
		return
	}

	// 从请求中获取注册表单数据
	username := r.FormValue("username")
	password := r.FormValue("password")

	// 验证表单数据是否有效

	if username == "" || password == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Missing required fields")
		return
	}

	if utf8RuneCountInString(username) > 10 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Username cannot exceed 10 characters")
		return
	}

	if len(password) < 8 || len(password) > 100 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Password length should be between 8 and 100 characters")
		return
	}

	if !containsLetterAndNumber(password) || !containsLowerAndUpperCase(password) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Password must contain both letters and numbers, and have both lower and uppercase")
		return
	}

	// 进行注册逻辑的处理
	// 生成盐
	salt, err := generateSalt()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to generate salt")
		return
	}

	// 哈希密码
	hashedPassword := hashPassword(password, salt)
	logger.Debug("注册时生成盐:", salt)

	// 将用户数据存入数据库
	userID, err := saveUserToDB(username, hashedPassword, salt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to save user data")
		logger.Error("用户注册时出现错误:", err)
		return
	}

	// 返回用户唯一的自增ID
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Registration successful. UserID: %d", userID)

}

func utf8RuneCountInString(s string) int {
	return len([]rune(s))
}

func containsLetterAndNumber(s string) bool {
	match, _ := regexp.MatchString(`[a-zA-Z]+`, s)
	if !match {
		return false
	}
	match, _ = regexp.MatchString(`[0-9]+`, s)
	return match
}

func containsLowerAndUpperCase(s string) bool {
	return strings.ToLower(s) != s && strings.ToUpper(s) != s
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

		passwordHash, passwordSalt, err := getDBPasswordHash(userID)
		if err != nil {
			logger.Error("读取数据库密码哈希值失败", err)
		}
		logger.Debug("登录时读取盐:", passwordSalt)
		tryingPasswordHash := hashPassword(p.Password, passwordSalt)
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

func saveUserToDB(username, hashedPassword string, salt []byte) (int64, error) {
	query := "INSERT INTO userdatatable (userName, userPasswordHashValue, passwordSalt) VALUES (?, ?, ?)"
	result, err := db.Exec(query, username, hashedPassword, salt)
	if err != nil {
		return 0, err
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func getDBPasswordHash(userID int) (string, []byte, error) {
	query := "SELECT userPasswordHashValue, passwordSalt FROM userdatatable WHERE userID = ?"
	row := db.QueryRow(query, userID)

	var passwordHash string
	var salt []byte
	err := row.Scan(&passwordHash, &salt)
	if err != nil {
		if err == sql.ErrNoRows {
			// 用户不存在
			return "", nil, fmt.Errorf("User Not Found")
		}
		// 处理其他查询错误
		return "", nil, err
	}

	return passwordHash, salt, nil
}

func generateSalt() ([]byte, error) {
	salt := make([]byte, saltLength)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}

func hashPassword(password string, salt []byte) string {
	// 将盐与密码组合
	passwordWithSalt := append([]byte(password), salt...)

	// 使用 SHA-256 计算密码和盐的哈希值
	hash := sha256.New()
	hash.Write(passwordWithSalt)
	hashValue := hash.Sum(nil)

	// 返回哈希值的十六进制表示
	return hex.EncodeToString(hashValue)
}

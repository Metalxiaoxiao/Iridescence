package commandSystem

import (
	"bufio"
	"config"
	"dbUtils"
	"fmt"
	"github.com/gorilla/websocket"
	"httpService"
	"logger"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
	"websocketService"
)

type CommandHandler func(args []string)

var startTime time.Time

func init() {
	startTime = time.Now()
}

var commands = map[string]CommandHandler{
	"quit":       handleQuit,
	"status":     handleStatus,
	"kicktoken":  handleInvalidateToken,
	"userinfo":   handleUserInfo,
	"kick":       handleKickUser,
	"listusers":  handleListUsers,
	"listtokens": handleListTokens,
	"banuser":    handleBanUser,
	"unbanuser":  handleUnbanUser,
	"broadcast":  handleBroadcast,
}

func StartListening() {
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print(">> ")
			line, _ := reader.ReadString('\n')
			line = strings.TrimSpace(line)
			parts := strings.Split(line, " ")
			if handler, ok := commands[parts[0]]; ok {
				handler(parts[1:])
			} else {
				fmt.Println("未知的命令：", parts[0])
			}
		}
	}()
}

func handleQuit(args []string) {
	logger.Info("服务器正在关闭...")
	os.Exit(0)
}

func handleStatus(args []string) {
	logger.Info("当前服务器信息如下:")

	// 获取内存使用情况
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	memInMiB := float64(mem.Alloc) / 1024 / 1024

	// 获取CPU使用情况
	numCPU := runtime.NumCPU()

	// 获取运行时间
	upTime := time.Since(startTime)

	// 获取连接的用户数量
	numUsers := len(websocketService.Clients) // 假设你的 tokens 变量存储了所有连接的用户

	// 打印表头
	fmt.Printf("%-20s %-20s\n", "Metric", "Value")
	fmt.Printf("%-20s %-20s\n", "------", "-----")

	// 打印表格内容
	fmt.Printf("%-20s %.2f MiB\n", "Memory usage:", memInMiB)
	fmt.Printf("%-20s %d\n", "CPU:", numCPU)
	fmt.Printf("%-20s %v\n", "Uptime:", upTime)
	fmt.Printf("%-20s %d\n", "Number of connected users:", numUsers)
	fmt.Printf("%-20s %d\n", "Number of tokens issued:", len(httpService.Tokens))
}

func handleInvalidateToken(args []string) {
	if len(args) != 1 {
		fmt.Println("Usage: kicktoken [token]")
		return
	}

	token := args[0]
	if _, ok := httpService.Tokens[token]; ok {
		delete(httpService.Tokens, token)
		fmt.Println("Token invalidated:", token)
	} else {
		fmt.Println("Token not found:", token)
	}
}
func handleUserInfo(args []string) {
	if len(args) != 1 {
		fmt.Println("Usage: userinfo [userID]")
		return
	}

	userID, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("Invalid userID:", args[0])
		return
	}

	websocketService.ClientsLock.Lock()
	user, ok := websocketService.Clients[userID]
	websocketService.ClientsLock.Unlock()

	if !ok {
		fmt.Println("User not found:", userID)
		return
	}

	fmt.Printf("用户ID: %d\n", user.UserID)
	fmt.Printf("用户名: %s\n", user.UserName)
	fmt.Printf("用户头像: %s\n", user.UserAvatar)
	fmt.Printf("用户备注: %s\n", user.UserNote)
	fmt.Printf("用户权限: %d\n", user.UserPermission)
	fmt.Printf("用户好友列表: %s\n", string(user.UserFriendList))
}
func handleKickUser(args []string) {
	if len(args) != 1 {
		fmt.Println("Usage: kickuser [userID]")
		return
	}

	userID, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("Invalid userID:", args[0])
		return
	}

	websocketService.ClientsLock.Lock()
	user, ok := websocketService.Clients[userID]
	websocketService.ClientsLock.Unlock()

	if !ok {
		fmt.Println("User not found:", userID)
		return
	}

	err = user.Conn.Close()
	if err != nil {
		fmt.Println("Failed to disconnect user:", err)
		return
	}

	fmt.Println("User disconnected:", userID)
}
func handleListUsers(args []string) {
	websocketService.ClientsLock.Lock()
	defer websocketService.ClientsLock.Unlock()

	logger.Info("当前在线用户:")
	for id, user := range websocketService.Clients {
		fmt.Printf("用户ID: %d, 用户名: %s\n", id, user.UserName)
	}
}
func handleBroadcast(args []string) {
	if len(args) != 1 {
		fmt.Println("Usage: broadcast [message]")
		return
	}

	message := args[0]

	websocketService.ClientsLock.Lock()
	defer websocketService.ClientsLock.Unlock()

	for _, user := range websocketService.Clients {
		err := user.Conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			fmt.Println("Failed to send message to user:", user.UserID)
		}
	}

	fmt.Println("Broadcast message sent.")
}
func handleBanUser(args []string) {
	if len(args) != 1 {
		fmt.Println("Usage: banuser [userID]")
		return
	}

	userID, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("Invalid userID:", args[0])
		return
	}

	db := dbUtils.GetDBPtr()
	_, err = db.Exec("UPDATE userdatatable SET userPermission = ? WHERE userID = ?", config.PermissionBannedUser, userID)
	if err != nil {
		fmt.Println("Failed to ban user:", err)
		return
	}

	websocketService.ClientsLock.Lock()
	user, ok := websocketService.Clients[userID]
	if ok {
		user.UserPermission = config.PermissionBannedUser
	}
	websocketService.ClientsLock.Unlock()

	token, ok := httpService.UserToTokens[userID]
	if ok {
		tokenUser, _ := httpService.Tokens[*token]
		tokenUser.UserPermission = config.PermissionBannedUser
	}
	fmt.Println("User banned:", userID)
}
func handleUnbanUser(args []string) {
	if len(args) != 1 {
		fmt.Println("Usage: unbanuser [userID]")
		return
	}

	userID, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("Invalid userID:", args[0])
		return
	}

	db := dbUtils.GetDBPtr()
	_, err = db.Exec("UPDATE userdatatable SET userPermission = ? WHERE userID = ?", config.PermissionOrdinaryUser, userID)
	if err != nil {
		fmt.Println("Failed to unban user:", err)
		return
	}

	websocketService.ClientsLock.Lock()
	user, ok := websocketService.Clients[userID]
	if ok {
		user.UserPermission = config.PermissionOrdinaryUser
	}
	websocketService.ClientsLock.Unlock()

	token, ok := httpService.UserToTokens[userID]
	if ok {
		tokenUser, _ := httpService.Tokens[*token]
		tokenUser.UserPermission = config.PermissionOrdinaryUser
	}

	fmt.Println("User unbanned:", userID)
}
func handleListTokens(args []string) {
	fmt.Println("当前有效的tokens:")
	for token, user := range httpService.Tokens {
		// 检查token是否已经过期
		if httpService.CheckTokenExpiry(token) {
			fmt.Printf("Token: %s, 用户ID: %d, 权限等级: %d\n", token, user.UserID, user.UserPermission)
		}
	}
}

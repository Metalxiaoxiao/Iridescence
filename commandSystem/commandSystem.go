package commandSystem

import (
	"bufio"
	"fmt"
	"httpService"
	"logger"
	"os"
	"runtime"
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
	"quit":     handleQuit,
	"status":   handleStatus,
	"kickkick": handleInvalidateToken,
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
				fmt.Println(">>", parts[0])
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
		fmt.Println("Usage: kick [token]")
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

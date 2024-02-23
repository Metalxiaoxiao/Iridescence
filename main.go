package main

import (
	"config"
	"dbUtils"
	fileserver "filesystem"
	"fmt"
	"hashUtils"
	"httpService"
	"logger"
	"net/http"
	"time"
	wsService "websocketService"

	_ "github.com/go-sql-driver/mysql"
)

const (
	_VERSION = "0.0.1"
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
	fmt.Println("                     VERSION:", _VERSION)
	logger.SetLogLevel(logger.LogLevel(confData.LogLevel))

	var (
		_PROT = confData.Port
	)

	logger.Info("Server running at port:", _PROT)

	if confData.DataBaseSettings.Account == "default_account" {
		logger.Error("配置文件中数据库信息未修改，请修改后再启动程序")
		time.Sleep(time.Second * 5)
		return
	}
	logger.Info("Trying to connect to MYSQL...")
	dbUtils.DbInit(confData)
	hashUtils.LoadConfig(confData)
	httpService.LoadConfig(confData)
	wsService.LoadConfig(confData)

	db := dbUtils.GetDBPtr()
	wsService.LoadDB(db)

	logger.Info("服务器启动成功！")
	http.HandleFunc(confData.Rotes.WebSocketServiceRote, wsService.HandleWebSocket)
	http.HandleFunc(confData.Rotes.RegisterServiceRote, httpService.HandleRegister)
	http.HandleFunc(confData.Rotes.LoginServiceRote, httpService.HandleLogin)
	http.HandleFunc(confData.Rotes.RequestServiceRote, httpService.HandleRequest)
	http.HandleFunc(confData.Rotes.UploadServiceRote, fileserver.HandleFileUpload)
	http.HandleFunc(confData.Rotes.DownloadServiceRote, fileserver.HandleFileDownload)
	logger.Error(http.ListenAndServe(":"+_PROT, nil))
}

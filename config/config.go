package config

import (
	"encoding/json"
	"fmt"
	"io"
	"logger"
	"os"
)

type Config struct {
	LogLevel         int    `json:"logLevel"`
	Port             string `json:"port"`
	DataBaseSettings struct {
		Address  string `json:"address"`
		Account  string `json:"account"`
		Password string `json:"password"`
	}
	RegisterServiceRote    string `json:"registerRote"`
	LoginServiceRote       string `json:"loginRote"`
	WebSocketServiceRote   string `json:"wsRote"`
	UploadServiceRote      string `json:"uploadRote"`
	DownloadServiceRote    string `json:"downloadRote"`
	SaltLength             int
	TokenLength            int
	AuthorizedServerTokens []string `json:"authorizedServerTokens"`
	TokenExpiryHours       float64
}

// LoadConfig 从指定的文件路径加载配置文件，如果文件不存在则创建并写入默认配置
func LoadConfig(filename string) (Config, error) {
	var config Config

	// 尝试打开配置文件
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return config, fmt.Errorf("无法打开配置文件: %v", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			logger.Error("打开配置文件失败", err)
		}
	}(file)

	// 获取文件的状态信息
	fileInfo, err := file.Stat()
	if err != nil {
		return config, fmt.Errorf("无法获取文件信息: %v", err)
	}

	// 检查文件是否为空
	if fileInfo.Size() == 0 {
		logger.Warn("配置文件不存在，写入默认配置")
		// 写入默认配置
		defaultConfig := getDefaultConfig()
		err = writeConfigToFile(file, defaultConfig)
		if err != nil {
			return config, fmt.Errorf("无法写入默认配置: %v", err)
		}
	}

	// 读取配置文件内容
	configBytes, err := io.ReadAll(file)
	if err != nil {
		return config, fmt.Errorf("无法读取配置文件: %v", err)
	}

	// 解析配置文件内容为 Config 结构体
	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		return config, fmt.Errorf("无法解析配置文件: %v", err)
	}

	return config, nil
}

// WriteConfig 将配置写入指定的文件路径
func WriteConfig(filename string, config Config) error {
	// 尝试打开配置文件
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("无法打开配置文件: %v", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			logger.Error("关闭配置文件失败", err)
		}
	}(file)

	// 写入配置文件
	err = writeConfigToFile(file, config)
	if err != nil {
		return fmt.Errorf("无法写入配置文件: %v", err)
	}

	return nil
}

// writeConfigToFile 将配置写入文件
func writeConfigToFile(file *os.File, config Config) error {
	// 将配置结构体转换为 JSON 格式
	configBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("无法转换配置为 JSON: %v", err)
	}

	// 清空文件内容
	err = file.Truncate(0)
	if err != nil {
		return fmt.Errorf("无法清空文件内容: %v", err)
	}

	// 将配置写入文件
	_, err = file.WriteAt(configBytes, 0)
	if err != nil {
		return fmt.Errorf("无法写入配置文件: %v", err)
	}

	return nil
}

// getDefaultConfig 返回默认配置
func getDefaultConfig() Config {
	// 设置默认配置
	defaultConfig := Config{
		LogLevel: 2,
		Port:     "8080",
		DataBaseSettings: struct {
			Address  string `json:"address"`
			Account  string `json:"account"`
			Password string `json:"password"`
		}{
			Address:  "localhost:3306",
			Account:  "default_account",
			Password: "default_password",
		},
		RegisterServiceRote:    "/register",
		LoginServiceRote:       "/login",
		WebSocketServiceRote:   "/ws",
		UploadServiceRote:      "/upload",
		DownloadServiceRote:    "/download",
		SaltLength:             8,
		TokenLength:            256,
		TokenExpiryHours:       24.00,
		AuthorizedServerTokens: []string{"token1", "token2", "token3"},
	}

	return defaultConfig
}

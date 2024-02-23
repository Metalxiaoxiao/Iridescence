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
	Rotes struct {
		RegisterServiceRote  string `json:"registerRote"`
		RequestServiceRote   string `json:"requestRote"`
		LoginServiceRote     string `json:"loginRote"`
		WebSocketServiceRote string `json:"wsRote"`
		UploadServiceRote    string `json:"uploadRote"`
		DownloadServiceRote  string `json:"downloadRote"`
	}
	WebsocketConnBufferSize int      `json:"WebsocketConnBufferSize"`
	SaltLength              int      `json:"saltLength"`
	TokenLength             int      `json:"tokenLength"`
	AuthorizedServerTokens  []string `json:"authorizedServerTokens"`
	TokenExpiryHours        float64  `json:"TokenExpiryHours"`
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
	CheckConfigIntegrity(filename, &config)
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

// CheckConfigIntegrity 检查配置文件的完整性，并自动补充缺少的字段
func CheckConfigIntegrity(filename string, config *Config) {
	// 将配置转换为 map
	configMap := structToMap(*config)

	// 获取默认配置
	defaultConfig := getDefaultConfig()

	// 将默认配置转换为 map
	defaultConfigMap := structToMap(defaultConfig)

	// 检查缺少的字段并自动补充
	checkAndFillMissingFields(configMap, defaultConfigMap)

	// 将 map 转换回 Config 结构体
	mapToStruct(configMap, config)

	// 将补充后的配置写入文件
	err := WriteConfig(filename, *config)
	if err != nil {
		logger.Error("无法写入配置文件: ", err)
	}
}

// structToMap 将结构体转换为 map
func structToMap(s interface{}) map[string]interface{} {
	bytes, _ := json.Marshal(s)
	var result map[string]interface{}
	err := json.Unmarshal(bytes, &result)
	if err != nil {
		return nil
	}
	return result
}

// mapToStruct 将 map 转换为 Config 结构体
func mapToStruct(m map[string]interface{}, s *Config) {
	bytes, _ := json.Marshal(m)
	err := json.Unmarshal(bytes, &s)
	if err != nil {
		return
	}
}

// isEmptyValue 检查值是否为零值
func isEmptyValue(v interface{}) bool {
	switch v := v.(type) {
	case bool:
		return !v
	case int, int8, int16, int32, int64:
		return v == 0
	case uint, uint8, uint16, uint32, uint64, uintptr:
		return v == 0
	case float32, float64:
		return v == 0
	case string:
		return v == ""
	case []interface{}:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	case nil:
		return true
	default:
		return false
	}
}

// checkAndFillMissingFields 检查并填充缺少的字段
func checkAndFillMissingFields(target, source map[string]interface{}) {
	for key, sourceValue := range source {
		if targetValue, ok := target[key]; !ok || isEmptyValue(targetValue) {
			target[key] = sourceValue
		} else {
			// 如果字段是嵌套的 map，递归检查
			if targetMap, ok := targetValue.(map[string]interface{}); ok {
				sourceMap := sourceValue.(map[string]interface{})
				checkAndFillMissingFields(targetMap, sourceMap)
			}
		}
	}
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
		Rotes: struct {
			RegisterServiceRote  string `json:"registerRote"`
			RequestServiceRote   string `json:"requestRote"`
			LoginServiceRote     string `json:"loginRote"`
			WebSocketServiceRote string `json:"wsRote"`
			UploadServiceRote    string `json:"uploadRote"`
			DownloadServiceRote  string `json:"downloadRote"`
		}{
			RegisterServiceRote:  "/register",
			RequestServiceRote:   "/request",
			LoginServiceRote:     "/login",
			WebSocketServiceRote: "/ws",
			UploadServiceRote:    "/upload",
			DownloadServiceRote:  "/download",
		},
		SaltLength:              8,
		TokenLength:             256,
		TokenExpiryHours:        24.00,
		WebsocketConnBufferSize: 2048,
		AuthorizedServerTokens:  []string{"token1", "token2", "token3"},
	}

	return defaultConfig
}

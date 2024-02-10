package jsonprovider

import (
	"logger"

	jsoniter "github.com/json-iterator/go"
)

func ParseJSON(jsonByte []byte, target interface{}) {
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	err := json.Unmarshal(jsonByte, target)
	if err != nil {
		logger.Error("JSON解码错误", err)
	}
}
func StringifyJSON(data interface{}) []byte {
	// 创建 JSON 编码器
	json := jsoniter.ConfigCompatibleWithStandardLibrary

	// 编码为 JSON 数据
	jsonData, err := json.Marshal(data)
	if err != nil {
		logger.Error("JSON编码错误", err)
		return nil
	}

	return jsonData
}

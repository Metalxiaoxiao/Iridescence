package jsonprovider

import (
	"io"
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
func WriteJSONToWriter(writer io.Writer, data interface{}) {
	// 创建 JSON 编码器
	json := jsoniter.ConfigCompatibleWithStandardLibrary

	// 编码为 JSON 数据并写入到 io.Writer 中
	err := json.NewEncoder(writer).Encode(data)
	if err != nil {
		logger.Error("JSON编码错误", err)
	}
}

func SdandarlizeJSON_byte(command string, content interface{}) []byte {
	return StringifyJSON(SdandarlizeJSON(command, content))
}

func SdandarlizeJSON(command string, content interface{}) interface{} {
	var res = StandardJSONPack{
		Command: command,
		Content: content,
	}
	logger.Debug("服务器回发包：", res)
	return res
}

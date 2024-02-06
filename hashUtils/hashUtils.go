package hashUtils

import (
	"config"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

var confData config.Config

func LoadConfig(conf config.Config) {
	confData = conf
}

func GenerateSalt() ([]byte, error) {
	salt := make([]byte, confData.SaltLength)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}

func HashPassword(password string, salt []byte) string {
	// 将盐与密码组合
	passwordWithSalt := append([]byte(password), salt...)

	// 使用 SHA-256 计算密码和盐的哈希值
	hash := sha256.New()
	hash.Write(passwordWithSalt)
	hashValue := hash.Sum(nil)

	// 返回哈希值的十六进制表示
	return hex.EncodeToString(hashValue)
}

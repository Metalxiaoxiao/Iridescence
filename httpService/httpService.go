package httpService

import (
	"config"
	"dbUtils"
	"fmt"
	"hashUtils"
	"logger"
	"net/http"
	"regexp"
	"strings"
)

var configData config.Config

func LoadConfig(conf config.Config) {
	configData = conf
}

// 注册逻辑
func HandleRegister(w http.ResponseWriter, r *http.Request) {
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
	salt, err := hashUtils.GenerateSalt()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to generate salt")
		return
	}

	// 哈希密码
	hashedPassword := hashUtils.HashPassword(password, salt)
	logger.Debug("注册时生成盐:", salt)

	// 将用户数据存入数据库
	userID, err := dbUtils.SaveUserToDB(username, hashedPassword, salt)
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

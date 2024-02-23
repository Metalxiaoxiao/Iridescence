package httpService

import (
	"config"
	"dbUtils"
	"encoding/json"
	"fmt"
	"hashUtils"
	"logger"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var configData config.Config

func LoadConfig(conf config.Config) {
	configData = conf
}
db := dbUtils.GetDBPtr()

type User struct {
	UserID         int
	Expiry         time.Time
	Username       string
	UserName       string          `json:"userName"`
	UserAvatar     string          `json:"userAvatar"`
	UserNote       string          `json:"userNote"`
	UserPermission uint            `json:"userPermission"`
	UserFriendList json.RawMessage `json:"userFriendList"`
}

var tokens map[string]User = make(map[string]User)

// 注册逻辑
func HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "不允许GET请求，请使用POST重新请求")
		return
	}

	// 从请求中获取注册表单数据
	username := r.FormValue("userName")
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

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "不允许GET请求，请使用POST重新请求")
		return
	}

	// 从请求中获取登录表单数据
	userID := r.FormValue("userId")
	password := r.FormValue("password")

	// 验证表单数据是否有效
	if userID == "" || password == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Missing required fields")
		return
	}
	userIDint,_ := strconv.Atoi(userID)

	passwordHash, passwordSalt, err := dbUtils.GetDBPasswordHash(userIDint)
	if err != nil {
		logger.Error("读取数据库密码哈希值失败", err)
	}
	logger.Debug("登录时读取盐:", passwordSalt)
	tryingPasswordHash := hashUtils.HashPassword(password, passwordSalt)
	logger.Debug("尝试哈希", tryingPasswordHash, "实际哈希", passwordHash)
	if tryingPasswordHash == passwordHash {

		// 从数据库中获取用户信息
		var username, userAvatar, userNote string
		var userPermission uint
		var userFriendList json.RawMessage
		err := db.QueryRow("SELECT userName, userAvatar, userNote, userPermission, userFriendList FROM userdatatable WHERE userID = ?", userID).Scan(&username, &userAvatar, &userNote, &userPermission, &userFriendList)
		if err != nil {
			logger.Error("获取用户数据失败:", err)
			return
		}
		var user = &User{}
		// 生成token
		token, err := hashUtils.GenerateRandomToken()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Failed to generate token")
			return
		}

		// 设置token的失效时间为24小时后
		expiry := time.Now().Add(24 * time.Hour)
		user.Expiry = expiry

		// 将token和用户信息存入内存
		tokens[token] = user

		// 返回token给用户
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Login successful. Token: %s", token)
		logger.Debug("用户", userID, "登录成功")
	}
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
func CheckTokenExpiry(token string) bool {
	user, ok := tokens[token]
	if !ok {
		// Token不存在
		return false
	}

	// 检查Token是否已经过期
	if time.Now().After(user.Expiry) {
		// Token已经过期
		return false
	}

	// Token没有过期
	return true
}

package httpService

import (
	"config"
	"dbUtils"
	"fmt"
	"hashUtils"
	"io"
	jsonprovider "jsonProvider"
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
	// 遍历 authorizedServerTokens 并将它们添加到 Tokens map 中
	for _, token := range configData.AuthorizedServerTokens {
		user := User{
			UserID:         int(time.Now().UnixNano()),
			UserPermission: config.PermissionServer,
			UserName:       "OtherServer",
			TokenExpiry:    time.Now().Add(1024 * time.Hour),
		}
		Tokens[token] = user
	}
}

type User jsonprovider.User

var Tokens map[string]User = make(map[string]User)

func fmtPrintF(io io.Writer, content string, a ...any) {
	var err error
	if a == nil {
		_, err = fmt.Fprintf(io, content)
	} else {
		_, err = fmt.Fprintf(io, content, a)
	}

	if err != nil {
		logger.Error("流输出错误", err)
		return
	}
}

// 注册逻辑
func HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmtPrintF(w, "不允许GET请求，请使用POST重新请求")
		return
	}

	// 从请求中获取注册表单数据
	username := r.FormValue("userName")
	password := r.FormValue("password")

	// 验证表单数据是否有效

	if username == "" || password == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmtPrintF(w, "缺少参数")
		return
	}

	if utf8RuneCountInString(username) > 10 {
		w.WriteHeader(http.StatusBadRequest)
		fmtPrintF(w, "Username cannot exceed 10 characters")
		return
	}

	if len(password) < 8 || len(password) > 100 {
		w.WriteHeader(http.StatusBadRequest)
		fmtPrintF(w, "Password length should be between 8 and 100 characters")
		return
	}

	if !containsLetterAndNumber(password) || !containsLowerAndUpperCase(password) {
		w.WriteHeader(http.StatusBadRequest)
		fmtPrintF(w, "Password must contain both letters and numbers, and have both lower and uppercase")
		return
	}

	// 进行注册逻辑的处理
	// 生成盐
	salt, err := hashUtils.GenerateSalt()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmtPrintF(w, "密码加盐时出错")
		return
	}

	// 哈希密码
	hashedPassword := hashUtils.HashPassword(password, salt)
	logger.Debug("注册时生成盐:", salt)

	// 将用户数据存入数据库
	userID, err := dbUtils.SaveUserToDB(username, hashedPassword, salt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmtPrintF(w, "保存用户信息时出错")
		logger.Error("用户注册时出现错误:", err)
		return
	}

	// 返回用户唯一的自增ID
	w.WriteHeader(http.StatusOK)
	fmtPrintF(w, strconv.FormatInt(userID, 10))

}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmtPrintF(w, "不允许GET请求，请使用POST重新请求")
		return
	}

	// 从请求中获取登录表单数据
	userID := r.FormValue("userId")
	password := r.FormValue("password")

	// 验证表单数据是否有效
	if userID == "" || password == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmtPrintF(w, "缺少参数")
		return
	}
	userIDint, _ := strconv.Atoi(userID)

	passwordHash, passwordSalt, err := dbUtils.GetDBPasswordHash(userIDint)
	if err != nil {
		logger.Error("读取数据库密码哈希值失败", err)
	}
	logger.Debug("登录时读取盐:", passwordSalt)
	tryingPasswordHash := hashUtils.HashPassword(password, passwordSalt)
	logger.Debug("尝试哈希", tryingPasswordHash, "实际哈希", passwordHash)
	if tryingPasswordHash == passwordHash {
		res, _ := dbUtils.GetUserFromDB(userIDint)
		var user = User{
			UserID:         userIDint,
			UserAvatar:     res.UserAvatar,
			UserNote:       res.UserNote,
			UserPermission: res.UserPermission,
			UserFriendList: res.UserFriendList,
			UserName:       res.UserName,
		}
		// 生成token
		token, err := hashUtils.GenerateRandomToken()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmtPrintF(w, "无法生成Token")
			return
		}

		// 设置token的失效时间为24小时后
		expiry := time.Now().Add(time.Duration(configData.TokenExpiryHours * float64(time.Hour)))
		user.TokenExpiry = expiry

		// 将token和用户信息存入内存
		Tokens[token] = user

		// 返回token给用户
		w.WriteHeader(http.StatusOK)
		fmtPrintF(w, token)
		logger.Debug("用户", userID, "登录成功")
	}
}

// HandleRequest 处理查询用户信息的HTTP请求
func HandleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmtPrintF(w, "不允许GET请求，请使用POST重新请求")
		return
	}

	// 解析请求体
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmtPrintF(w, "Failed to parse form")
		return
	}

	// 从请求中获取token和用户ID
	token := r.FormValue("token")
	command := r.FormValue("command")

	user, ok := Tokens[token]

	// 验证token是否有效
	if !CheckTokenExpiry(token) || ok == false {
		w.WriteHeader(http.StatusUnauthorized)
		fmtPrintF(w, "Invalid token")
		return
	}

	// 验证用户ID是否有效
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmtPrintF(w, "Invalid userID")
		return
	}
	if user.UserPermission >= config.PermissionServer {
		//Server命令
		switch command {
		case "verifyToken":
			targetToken := r.FormValue("targetToken")
			logger.Debug("远端服务器尝试验证用户token", targetToken)
			targetUser, ok := Tokens[targetToken]

			// 验证token是否有效
			if !CheckTokenExpiry(token) || ok == false {
				w.WriteHeader(http.StatusUnauthorized)
				fmtPrintF(w, "Invalid token")
				return
			}
			w.WriteHeader(http.StatusOK)
			jsonprovider.WriteJSONToWriter(w, targetUser)
		default:
			w.WriteHeader(http.StatusBadRequest)
			fmtPrintF(w, "未知的命令")
		}
	}

	switch command {
	case "getUserData":
		// 从数据库中获取用户信息
		targetUser, err := dbUtils.GetUserFromDB(user.UserID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmtPrintF(w, "Failed to get user data")
			return
		}

		// 发送响应
		w.WriteHeader(http.StatusOK)
		jsonprovider.WriteJSONToWriter(w, targetUser)
	case "getUserDataByID":
		// 从数据库中获取用户信息
		targetUserID := r.FormValue("target")
		targetUserIDint, _ := strconv.Atoi(targetUserID)
		targetUser, err := dbUtils.GetUserFromDB(targetUserIDint)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmtPrintF(w, "Failed to get user data")
			return
		}

		// 发送响应
		w.WriteHeader(http.StatusOK)
		jsonprovider.WriteJSONToWriter(w, targetUser)
	default:
		w.WriteHeader(http.StatusBadRequest)
		fmtPrintF(w, "未知的命令")
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
	user, ok := Tokens[token]
	if !ok {
		// Token不存在
		return false
	}

	// 检查Token是否已经过期
	if time.Now().After(user.TokenExpiry) {
		// Token已经过期
		return false
	}

	// Token没有过期
	return true
}

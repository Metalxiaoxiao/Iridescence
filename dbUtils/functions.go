package dbUtils

import (
	"database/sql"
	"fmt"
	"logger"
	"time"
)

func SaveUserToDB(username, hashedPassword string, salt []byte) (int64, error) {
	UseDB(db, _BasicChatDBName)
	query := "INSERT INTO userdatatable (userName, userPasswordHashValue, passwordSalt) VALUES (?, ?, ?)"
	result, err := db.Exec(query, username, hashedPassword, salt)
	if err != nil {
		return 0, err
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return userID, nil
}

// SaveOfflineMessageToDB 返回messageID
func SaveOfflineMessageToDB(userID int, recipientID int, messageContent string, messageType int) (int, error) {
	insertQuery := "INSERT INTO offlinemessages (senderID,receiverID,messageBody,time,messageType) VALUES (?,?,?,?,?)"
	timestamp := time.Now().UnixNano() //纳秒事件戳
	result, err := db.Exec(insertQuery, userID, recipientID, messageContent, timestamp, messageType)
	if err != nil {
		logger.Error("保存用户离线消息时出现错误", err)
		return 0, err
	}

	messageID, err := result.LastInsertId()
	if err != nil {
		logger.Error("获取插入消息的ID时出现错误", err)
		return 0, err
	}

	return int(messageID), nil
}

func GetDBPasswordHash(userID int) (string, []byte, error) {
	UseDB(db, _BasicChatDBName)
	query := "SELECT userPasswordHashValue, passwordSalt FROM userdatatable WHERE userID = ?"
	row := db.QueryRow(query, userID)

	var passwordHash string
	var salt []byte
	err := row.Scan(&passwordHash, &salt)
	if err != nil {
		if err == sql.ErrNoRows {
			// 用户不存在
			return "", nil, fmt.Errorf("找不到用户")
		}
		// 处理其他查询错误
		return "", nil, err
	}

	return passwordHash, salt, nil
}
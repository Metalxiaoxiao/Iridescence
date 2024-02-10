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

func SaveOfflineMessageToDB(userID int, recipientID int, messageContent string, messageType int) {
	insertQuery := "INSERT INTO offlinemessages (senderID,receiverID,messageBody,time,messageType) VALUES (?,?,?,?,?)"
	timestamp := time.Now()
	_, err := db.Exec(insertQuery, userID, recipientID, messageContent, timestamp, 0)
	if err != nil {
		logger.Error("保存用户离线消息时出现错误", err)
	}
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

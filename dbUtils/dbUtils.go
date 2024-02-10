package dbUtils

import (
	"config"
	"database/sql"
	"logger"
	"strconv"
)

var (
	_BasicChatDBName = "basic_chat_base"
)
var db *sql.DB

func UseDB(db *sql.DB, DBname string) {
	_, err := db.Exec("USE " + DBname)
	if err != nil {
		logger.Error("操作 ", DBname, " 数据库时出现错误:", err)
	}
}

func CheckTableExistence(db *sql.DB, DBname string, tableName string) int {
	query := "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ? AND table_name = ?"
	var tablecount int
	err := db.QueryRow(query, DBname, tableName).Scan(&tablecount)
	if err != nil {
		logger.Error("Failed to check table existence:", err)
	}
	return tablecount
}

func GetDBPtr() *sql.DB {
	return db
}

func DbInit(confData config.Config) {
	db, err := sql.Open("mysql", confData.DataBaseSettings.Account+":"+confData.DataBaseSettings.Password+"@tcp("+confData.DataBaseSettings.Address+")/")
	if err != nil {
		logger.Error(err)
	} else {
		logger.Info("Connected to MYSQL successfully")
	}
	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS " + _BasicChatDBName)
	if err != nil {
		logger.Error("创建 ", _BasicChatDBName, " 数据库时出现错误:", err)
	}
	UseDB(db, _BasicChatDBName)
	// 如果表不存在，则创建表
	if CheckTableExistence(db, _BasicChatDBName, "userdatatable") == 0 {
		UseDB(db, _BasicChatDBName)
		logger.Warn("找不到用户数据表，自动创建")
		createTable := `
		CREATE TABLE userdatatable (
			userID int unsigned NOT NULL AUTO_INCREMENT,
			userName varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
			userAvatar text,
			userNote varchar(255) DEFAULT NULL,
			userPermission int unsigned DEFAULT 4,
			userFriendList json DEFAULT NULL,
			userGroupList json DEFAULT NULL,
			userHomePageData json DEFAULT NULL,
			userSettings json DEFAULT NULL,
			userPasswordHashValue text,
			passwordSalt BINARY(` + strconv.Itoa(confData.SaltLength) + `),
			PRIMARY KEY (userID)
		  ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
		`
		_, err := db.Exec(createTable)
		if err != nil {
			logger.Error("Failed to create table:", err)
		}
	}
	if CheckTableExistence(db, _BasicChatDBName, "groupdatatable") == 0 {
		UseDB(db, _BasicChatDBName)
		logger.Warn("找不到群聊数据表，自动创建")
		createTable := `
			CREATE TABLE groupdatatable (
				groupID int NOT NULL AUTO_INCREMENT,
				groupName varchar(255) NOT NULL,
				groupAvatar varchar(255) DEFAULT NULL,
				groupExplaination text NOT NULL,
				groupMaster int DEFAULT NULL,
				groupMembers json DEFAULT NULL,
				groupSettings json DEFAULT NULL,
				PRIMARY KEY (groupID)
			  ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
		`
		_, err := db.Exec(createTable)
		if err != nil {
			logger.Error("Failed to create table:", err)
		}
	}
	if CheckTableExistence(db, _BasicChatDBName, "offlinegroupmessages") == 0 {
		UseDB(db, _BasicChatDBName)
		logger.Warn("找不到离线群消息数据表，自动创建")
		createTable := `CREATE TABLE offlinegroupmessages (
    			senderID int unsigned DEFAULT NULL,
				receiverID int unsigned NOT NULL,
				messageID int unsigned DEFAULT NULL,
				time datetime DEFAULT NULL,
				messageBody text DEFAULT NULL,
				messageType smallint unsigned DEFAULT NULL,
				PRIMARY KEY (receiverID) USING BTREE,
				KEY idx_senderID (senderID)
			  ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;`
		_, err := db.Exec(createTable)
		if err != nil {
			logger.Error("Failed to create table:", err)
		}
	}
	if CheckTableExistence(db, _BasicChatDBName, "offlinemessages") == 0 {
		UseDB(db, _BasicChatDBName)
		logger.Warn("找不到离线消息数据表，自动创建")
		createTable := `CREATE TABLE offlinemessages (
    			senderID int unsigned DEFAULT NULL,
				receiverID int unsigned NOT NULL,
				messageID int unsigned DEFAULT NULL,
				time datetime DEFAULT NULL,
				messageBody text DEFAULT NULL,
				messageType smallint unsigned DEFAULT NULL,
				PRIMARY KEY (receiverID) USING BTREE,
				KEY idx_senderID (senderID)
			  ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;`
		_, err := db.Exec(createTable)
		if err != nil {
			logger.Error("Failed to create table:", err)
		}
	}
}

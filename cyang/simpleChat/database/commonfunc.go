package database

import (
	"database/sql"
	"encoding/base64"
	"log"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func OpenDatabase() *sql.DB {
	db, err := sql.Open("mysql", "root:5kasile123@/simpleChat")
	err = db.Ping()
	if err != nil {
		log.Println(err)
		return db
	}
	return db
}

func CloseDatabase(db *sql.DB) {
	db.Close()
}

type User struct {
	Id       int
	Username string
	Password string
}

type FriendRelationship struct {
	Id       int
	Userid   int
	Friendid int
}

type ChatInfo struct {
	Sender   string
	Reciever string
	Time     string
	Body     string
}

//用于加密字符串使得字符传输过程中没有"|"和";"符号
const (
	base64Table        = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	hashFunctionHeader = "zh.ife.iya"
	hashFunctionFooter = "09.O25.O20.78"
)

var coder = base64.NewEncoding(base64Table)

/**
 * base64加密
 */
func Base64Encode(str string) string {
	var src []byte = []byte(hashFunctionHeader + str + hashFunctionFooter)
	return string([]byte(coder.EncodeToString(src)))
}

/**
 * base64解密
 */
func Base64Decode(str string) (string, error) {
	var src []byte = []byte(str)
	by, err := coder.DecodeString(string(src))
	return strings.Replace(strings.Replace(string(by), hashFunctionHeader, "", -1), hashFunctionFooter, "", -1), err
}

package database

import (
	"container/list"
	"database/sql"
	"log"
	"strings"

	"github.com/garyburd/redigo/redis"
	_ "github.com/go-sql-driver/mysql"
)

func FetchOrCreateChatInfoTable(db *sql.DB) {
	result, err := db.Query("select * from chatinfo")
	if err != nil {
		log.Println(err)
		stmt, err := db.Prepare("CREATE TABLE `chatinfo` (`id` smallint(6) NOT NULL AUTO_INCREMENT, `senderid` smallint(6) NOT NULL, `recieverid` smallint(6) NOT NULL, `datetime` char(64) NOT NULL, `body` text NOT NULL, PRIMARY KEY(`id`));")

		if err != nil {
			log.Println(err)
		}
		rs, err := stmt.Exec()
		if err != nil {
			log.Println(err)
		} else {
			log.Println("Table chatinfo is created! \n")
		}
		rs.LastInsertId()
		return
	}
	result.Close()
}

func FetchOrCreateOfflineChatInfoTable(db *sql.DB) {
	result, err := db.Query("select * from chatinfo_offline")
	if err != nil {
		log.Println(err)
		stmt, err := db.Prepare("CREATE TABLE `chatinfo_offline` (`id` smallint(6) NOT NULL AUTO_INCREMENT, `senderid` smallint(6) NOT NULL, `recieverid` smallint(6) NOT NULL, `datetime` char(64) NOT NULL, `body` text NOT NULL, PRIMARY KEY(`id`));")

		if err != nil {
			log.Println(err)
		}
		rs, err := stmt.Exec()
		if err != nil {
			log.Println(err)
		} else {
			log.Println("Table chatinfo_offline is created! \n")
		}
		rs.LastInsertId()
		return
	}
	result.Close()
}

func SaveChatMessage(sender *User, reciever *User, datetime string, body string, db *sql.DB) {
	stmt, err := db.Prepare("INSERT INTO chatinfo (senderid, recieverid, datetime, body) VALUES (?, ?, ?, ?);")
	if err != nil {
		log.Println(err)
	}

	_, err = stmt.Exec(sender.Id, reciever.Id, datetime, body)
	if err != nil {
		log.Println(err)
	}
	log.Printf("Message saved from user %s to user %s", sender.Username, reciever.Username)
}

func SaveOfflineChatMessage(sender *User, reciever *User, datetime, body string, db *sql.DB) {
	stmt, err := db.Prepare("INSERT INTO chatinfo_offline (senderid, recieverid, datetime, body) VALUES (?, ?, ?, ?);")
	if err != nil {
		log.Println(err)
	}

	_, err = stmt.Exec(sender.Id, reciever.Id, datetime, body)
	if err != nil {
		log.Println(err)
	}
	log.Printf("Offline message saved from user %s to user %s", sender.Username, reciever.Username)
}

func GetAllHistory(user1 *User, user2 *User, db *sql.DB) *list.List {
	user1id := user1.Id
	user2id := user2.Id
	var (
		id         int
		senderid   int
		recieverid int
		datetime   string
		body       string
	)
	results, err := db.Query("select * from chatinfo where (senderid = ? and recieverid = ?) or (senderid = ? and recieverid = ?)", user1id, user2id, user2id, user1id)
	if err != nil {
		log.Println(err)
	}
	chatInfos := list.New()
	defer results.Close()
	for results.Next() {
		err := results.Scan(&id, &senderid, &recieverid, &datetime, &body)
		if err != nil {
			log.Println(err)
		}
		senderName, _ := GetNameByID(senderid, db)
		recieverName, _ := GetNameByID(recieverid, db)
		chatInfo := new(ChatInfo)
		chatInfo.Sender = senderName
		chatInfo.Reciever = recieverName
		chatInfo.Time = datetime
		chatInfo.Body = body
		chatInfos.PushBack(chatInfo)
	}
	return chatInfos
}

func GetAllOfflineMessageSender(user *User, db *sql.DB) []*User {
	userid := user.Id
	var senderid int
	results, err := db.Query("select senderid from chatinfo_offline where recieverid = ?", userid)
	if err != nil {
		log.Println(err)
	}
	senders := []*User{}
	senderidMap := make(map[int]int)
	defer results.Close()
	for results.Next() {
		err := results.Scan(&senderid)
		if err != nil {
			log.Println(err)
		}
		senderidMap[senderid] = 1
	}
	for sid, value := range senderidMap {
		if value == 1 {
			sender := GetUserByID(sid, db)
			senders = append(senders, sender)
		}
	}
	return senders
}

func GetAllOfflineMessage(reciever *User, sender *User, db *sql.DB) *list.List {
	rid := reciever.Id
	sid := sender.Id
	var (
		id         int
		senderid   int
		recieverid int
		datetime   string
		body       string
	)
	results, err := db.Query("select * from chatinfo_offline where senderid = ? and recieverid = ?;", sid, rid)
	if err != nil {
		log.Println(err)
	}
	chatInfos := list.New()
	defer results.Close()
	for results.Next() {
		err := results.Scan(&id, &senderid, &recieverid, &datetime, &body)
		if err != nil {
			log.Println(err)
		}
		senderName, _ := GetNameByID(senderid, db)
		recieverName, _ := GetNameByID(recieverid, db)
		chatInfo := new(ChatInfo)
		chatInfo.Sender = senderName
		chatInfo.Reciever = recieverName
		chatInfo.Time = datetime
		chatInfo.Body = body
		chatInfos.PushBack(chatInfo)
	}
	return chatInfos
}

func DeleteOfflineMessage(reciever *User, sender *User, db *sql.DB) {
	rid := reciever.Id
	sid := sender.Id
	stmt, err := db.Prepare("DELETE FROM chatinfo_offline WHERE senderid = ? AND recieverid = ?;")
	if err != nil {
		log.Println(err)
	}
	_, err = stmt.Exec(sid, rid)
	if err != nil {
		log.Println(err)
	}
}

func SaveUnreadMessage(sender *User, reciever *User, datetime string, body string, rc redis.Conn) {
	rname := reciever.Username
	sname := sender.Username

	umkey := "Unread_" + sname + "_" + rname
	encodebody := Base64Encode(body)
	umvalue := datetime + "|" + encodebody
	rc.Do("RPUSH", umkey, umvalue)

	indexkey := "UnreadMsg_" + rname
	indexfield := sname
	rc.Do("HSETNX", indexkey, indexfield, umkey)

}

func GetAllUnreadMessageSender(user *User, db *sql.DB, rc redis.Conn) []*User {
	uname := user.Username

	indexkey := "UnreadMsg_" + uname
	values, err := redis.Values(rc.Do("HKEYS", indexkey))

	result := []*User{}
	if err != nil {
		log.Println(err)
		return result
	}

	for _, v := range values {
		sendername := string(v.([]byte))
		sender := GetUserByName(sendername, db)
		result = append(result, sender)
	}
	return result
}

func GetAllUnreadMessage(reciever *User, sender *User, rc redis.Conn) *list.List {
	rname := reciever.Username
	sname := sender.Username
	result := list.New()

	umkey := "Unread_" + sname + "_" + rname
	indexkey := "UnreadMsg_" + rname
	indexfield := sname

	values, err := redis.Values(rc.Do("LRANGE", umkey, 0, -1))
	if err != nil {
		log.Println(err)
		return result
	}
	for _, v := range values {
		message := string(v.([]byte))
		datetime := strings.Split(message, "|")[0]
		encodebody := strings.Split(message, "|")[1]
		tmp := new(ChatInfo)
		tmp.Sender = sname
		tmp.Reciever = rname
		tmp.Time = datetime
		tmp.Body = encodebody
		result.PushBack(tmp)
	}
	rc.Do("DEL", umkey)
	rc.Do("HDEL", indexkey, indexfield)

	return result
}

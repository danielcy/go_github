package database

import (
	"container/list"
	"database/sql"
	"log"

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

func SaveChatMessage(sender *User, reciever *User, datetime, body string, db *sql.DB) {
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

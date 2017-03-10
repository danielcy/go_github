package database

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func FetchOrCreateFriendTable(db *sql.DB) {
	result, err := db.Query("select * from friend")
	if err != nil {
		log.Println(err)
		stmt, err := db.Prepare("CREATE TABLE `friend` (`id` smallint(6) NOT NULL AUTO_INCREMENT, `userid` smallint(6) NOT NULL, friendid smallint(6) NOT NULL, PRIMARY KEY(`id`));")

		if err != nil {
			log.Println(err)
		}
		rs, err := stmt.Exec()
		if err != nil {
			log.Println(err)
		} else {
			log.Println("Table friend is created! \n")
		}
		rs.LastInsertId()
		return
	}
	result.Close()
}

func FetchOrCreateFriendApplyTable(db *sql.DB) {
	result, err := db.Query("select * from friendapply")
	if err != nil {
		log.Println(err)
		stmt, err := db.Prepare("CREATE TABLE `friendapply` (`id` smallint(6) NOT NULL AUTO_INCREMENT, `userid` smallint(6) NOT NULL, friendid smallint(6) NOT NULL, PRIMARY KEY(`id`));")

		if err != nil {
			log.Println(err)
		}
		rs, err := stmt.Exec()
		if err != nil {
			log.Println(err)
		} else {
			log.Println("Table friendapply is created! \n")
		}
		rs.LastInsertId()
		return
	}
	result.Close()
}

func ApplyFriend(user *User, friend *User, db *sql.DB) bool {
	userid := user.Id
	friendid := friend.Id

	if CheckFriendRelationship(user, friend, db) == true || CheckFriendApplyRelationship(user, friend, db) == true || CheckFriendApplyRelationship(friend, user, db) == true {
		return false
	}
	stmt, err := db.Prepare("INSERT INTO friendapply (userid, friendid) VALUES (?, ?);")
	if err != nil {
		log.Println(err)
	}
	_, err = stmt.Exec(userid, friendid)
	if err != nil {
		log.Println(err)
	}
	return true
}

func GetFriendList(user *User, db *sql.DB) []*User {
	userid := user.Id
	var friendid int
	results, err := db.Query("select friendid from friend where userid = ?", userid)
	if err != nil {
		log.Println(err)
	}
	friends := []*User{}
	defer results.Close()
	for results.Next() {
		err := results.Scan(&friendid)
		if err != nil {
			log.Println(err)
		}
		friend := GetUserByID(friendid, db)
		friends = append(friends, friend)
	}
	return friends
}

func GetFriendApplyList(user *User, db *sql.DB) []*User {
	userid := user.Id
	var friendid int
	results, err := db.Query("select userid from friendapply where friendid = ?", userid)
	if err != nil {
		log.Println(err)
	}
	friends := []*User{}
	defer results.Close()
	for results.Next() {
		err := results.Scan(&friendid)
		if err != nil {
			log.Println(err)
		}
		friend := GetUserByID(friendid, db)
		friends = append(friends, friend)
	}
	return friends
}

func ComfirmPendingFriend(user *User, friend *User, db *sql.DB) {
	userid := user.Id
	friendid := friend.Id
	stmt, err := db.Prepare("INSERT INTO friend (userid, friendid) VALUES (?, ?);")
	if err != nil {
		log.Println(err)
	}
	_, err = stmt.Exec(userid, friendid)
	if err != nil {
		log.Println(err)
	}
	_, err = stmt.Exec(friendid, userid)
	if err != nil {
		log.Println(err)
	}
	stmt, err = db.Prepare("DELETE FROM friendapply WHERE userid = ? AND friendid = ?;")
	if err != nil {
		log.Println(err)
	}
	_, err = stmt.Exec(friendid, userid)
	if err != nil {
		log.Println(err)
	}
}

func CheckFriendRelationship(user1 *User, user2 *User, db *sql.DB) bool {
	var id int
	err := db.QueryRow("select id from friend where userid = ? and friendid = ?", user1.Id, user2.Id).Scan(&id)
	if err != nil {
		return false
	}
	return true
}

func CheckFriendApplyRelationship(user1 *User, user2 *User, db *sql.DB) bool {
	var id int
	err := db.QueryRow("select id from friendapply where userid = ? and friendid = ?", user1.Id, user2.Id).Scan(&id)
	if err != nil {
		return false
	}
	return true
}

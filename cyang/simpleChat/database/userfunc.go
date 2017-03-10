package database

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func FetchOrCreateUserTable(db *sql.DB) {
	result, err := db.Query("select * from user")
	if err != nil {
		log.Println(err)
		stmt, err := db.Prepare("CREATE TABLE `user` (`userid` smallint(6) NOT NULL AUTO_INCREMENT, `username` char(16) NOT NULL, `password` char(16) NOT NULL, `status` tinyint default 0 NOT NULL, PRIMARY KEY(`userid`));")
		//CREATE TABLE `user` ( `userid` smallint(6) NOT NULL AUTO_INCREMENT, `password` char(16) NOT NULL, `username` char(16) NOT NULL, PRIMARY KEY(`userid`) ) ;

		if err != nil {
			log.Println(err)
		}
		rs, err := stmt.Exec()
		if err != nil {
			log.Println(err)
		} else {
			log.Println("Table user is created! \n")
		}
		rs.LastInsertId()
		return
	}
	result.Close()
}

func CheckUsername(username string, db *sql.DB) bool {
	var name string
	err := db.QueryRow("select username from user where username = ?", username).Scan(&name)
	if err != nil {
		return false
	}
	return true
}

func CheckPassword(username string, password string, db *sql.DB) bool {
	var pwd string
	err := db.QueryRow("select password from user where username = ?", username).Scan(&pwd)
	if err != nil {
		return false
	} else if password != pwd {
		return false
	}
	return true
}

func CheckStatus(username string, db *sql.DB) bool {
	var status int
	err := db.QueryRow("select status from user where username = ?", username).Scan(&status)
	if err != nil {
		return false
	} else if status == 0 {
		return false
	}
	return true
}

func AddUser(username string, password string, db *sql.DB) *User {
	stmt, err := db.Prepare("INSERT INTO user (username, password) VALUES (?, ?);")
	if err != nil {
		log.Println(err)
	}

	_, err = stmt.Exec(username, password)
	if err != nil {
		log.Println(err)
	}

	user := new(User)
	user.Id, _ = GetIDByName(username, db)
	user.Username = username
	user.Password = password

	return user
}

func SetStatusOn(user *User, db *sql.DB) {
	stmt, err := db.Prepare("UPDATE user SET status = 1 where username = ?")
	if err != nil {
		log.Println(err)
	}

	rs, err := stmt.Exec(user.Username)
	if err != nil {
		log.Println(err)
	}

	rs.LastInsertId()
}

func SetStatusOff(user *User, db *sql.DB) {
	stmt, err := db.Prepare("UPDATE user SET status = 0 where username = ?")
	if err != nil {
		log.Println(err)
	}

	rs, err := stmt.Exec(user.Username)
	if err != nil {
		log.Println(err)
	}

	rs.LastInsertId()
}

func SetAllStatusOff(db *sql.DB) {
	stmt, err := db.Prepare("UPDATE user SET status = 0;")
	if err != nil {
		log.Println(err)
	}

	_, err = stmt.Exec()
	if err != nil {
		log.Println(err)
	}
}

func GetIDByName(username string, db *sql.DB) (int, error) {
	var id int
	err := db.QueryRow("select userid from user where username = ?", username).Scan(&id)
	if err != nil {
		log.Println(err)
	}
	return id, err
}

func GetNameByID(userid int, db *sql.DB) (string, error) {
	var name string
	err := db.QueryRow("select username from user where userid = ?", userid).Scan(&name)
	if err != nil {
		log.Println(err)
	}
	return name, err
}

func GetUserByName(username string, db *sql.DB) *User {
	var id int
	err := db.QueryRow("select userid from user where username = ?", username).Scan(&id)
	if err != nil {
		log.Println(err)
	}
	var user *User = new(User)
	user.Id = id
	user.Username = username
	return user
}

func GetUserByID(userid int, db *sql.DB) *User {
	var name string
	err := db.QueryRow("select username from user where userid = ?", userid).Scan(&name)
	if err != nil {
		log.Println(err)
	}
	var user *User = new(User)
	user.Id = userid
	user.Username = name
	return user
}

func GetUserStatus(user *User, db *sql.DB) int {
	var status int
	err := db.QueryRow("select status from user where username = ?", user.Username).Scan(&status)
	if err != nil {
		return 0
	} else {
		return status
	}
}

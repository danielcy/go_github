package main

import (
	"database/sql"

	"github.com/cyang/simpleChat/database"
)

func LogInCheck(username string, password string, db *sql.DB) (bool, *database.User) {
	if database.CheckUsername(username, db) == true {
		if database.CheckPassword(username, password, db) == true {
			user := database.GetUserByName(username, db)
			return true, user
		}
	}
	return false, new(database.User)
}

func CheckUserExistance(username string, db *sql.DB) bool {
	if database.CheckUsername(username, db) == true {
		return true
	}
	return false
}

func AddUser(username string, password string, db *sql.DB) *database.User {
	return database.AddUser(username, password, db)
}

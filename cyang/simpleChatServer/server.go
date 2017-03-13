package main

import (
	"container/list"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/cyang/simpleChat/database"
	"github.com/garyburd/redigo/redis"
)

func checkError(err error, info string) (res bool) {
	if err != nil {
		fmt.Println(info + "  " + err.Error())
		return false
	}
	return true
}

func Handler(conn net.Conn, messages chan string, ipnamemap map[string]string, chattingTarget map[string]string, db *sql.DB, rc redis.Conn) {
	log.Println("connection is connected from ...", conn.RemoteAddr().String())

	buf := make([]byte, 1024)
	var user *database.User
	historyChatMsgs := list.New()
	for {
		lenght, err := conn.Read(buf)
		if checkError(err, "Connection") == false {
			if user != nil {
				database.SetStatusOff(user, db)
				delete(ipnamemap, user.Username)
				delete(chattingTarget, user.Username)
			}
			conn.Close()
			break
		}
		if lenght > 0 {
			buf[lenght] = 0
		}
		recieveStr := string(buf[0:lenght])
		head := strings.Split(recieveStr, "|")[0]

		switch head {
		case "LI": //Process Log In message.
			username := strings.Split(recieveStr, "|")[1]
			password := strings.Split(recieveStr, "|")[2]
			var exist bool
			exist, user = LogInCheck(username, password, db)
			isOnline := database.CheckStatus(username, db)
			if exist == false || isOnline == true {
				SendFailMessage(conn, head)
				continue
			}
			database.SetStatusOn(user, db)
			ipnamemap[user.Username] = conn.RemoteAddr().String()
			chattingTarget[user.Username] = ""
			log.Printf("Log in username: %s	userid: %d \n", user.Username, user.Id)
			returnMsg := conn.RemoteAddr().String() + "|SUCCESS|" + strconv.Itoa(user.Id)

			messages <- returnMsg

		case "CUM": //Process Check Unread Message message.
			offlineMessageSenders := database.GetAllUnreadMessageSender(user, db, rc)

			if len(offlineMessageSenders) != 0 {
				systemMsg := offlineMessageSenders[0].Username
				for i := 1; i < len(offlineMessageSenders); i++ {
					systemMsg = systemMsg + ", " + offlineMessageSenders[i].Username
				}
				systemMsg = conn.RemoteAddr().String() + "|SYSTEM|" + "You have some unread messages from " + systemMsg + ", Please remember to check. \n"
				log.Printf(systemMsg)
				messages <- systemMsg
			}

		case "SU": //Process Sign Up message.
			username := strings.Split(recieveStr, "|")[1]
			password := strings.Split(recieveStr, "|")[2]
			exist := CheckUserExistance(username, db)
			if exist == true {
				SendFailMessage(conn, head)
				continue
			}
			user = AddUser(username, password, db)
			database.SetStatusOn(user, db)
			ipnamemap[user.Username] = conn.RemoteAddr().String()
			chattingTarget[user.Username] = ""
			log.Printf("Sign up username: %s	userid: %d \n", user.Username, user.Id)
			returnMsg := conn.RemoteAddr().String() + "|SUCCESS|" + strconv.Itoa(user.Id)
			messages <- returnMsg

		case "AF": //Process Apply Friend message.
			friendname := strings.Split(recieveStr, "|")[1]
			exist := CheckUserExistance(friendname, db)
			if exist == false {
				SendFailMessage(conn, head)
				continue
			}
			friend := database.GetUserByName(friendname, db)
			log.Printf("User %s(id %d) sent an friend apply to user %s(id %d). \n", user.Username, user.Id, friend.Username, friend.Id)
			if database.ApplyFriend(user, friend, db) == false {
				SendFailMessage(conn, head)
				continue
			}
			returnMsg := conn.RemoteAddr().String() + "|SUCCESS"
			messages <- returnMsg
			// Also send a system message to the friend if he/she is online.
			friendAddr := ipnamemap[friend.Username]
			if friendAddr != "" {
				sendMsg := friendAddr + "|SYSTEM|" + user.Username + " sends you a friend application."
				messages <- sendMsg
			}

		case "VRL": //Process View Request List message.
			requestFriends := database.GetFriendApplyList(user, db)
			returnMsg := conn.RemoteAddr().String()
			size := len(requestFriends)
			for i := 0; i < size; i++ {
				returnMsg = returnMsg + "|" + strconv.Itoa(requestFriends[i].Id) + ";" + requestFriends[i].Username
			}
			messages <- returnMsg

		case "CR": //Process Confirm Request message.
			friendid, _ := strconv.Atoi(strings.Split(recieveStr, "|")[1])
			friend := database.GetUserByID(friendid, db)
			if database.CheckFriendApplyRelationship(friend, user, db) == false {
				SendFailMessage(conn, head)
				continue
			}
			database.ComfirmPendingFriend(user, friend, db)
			returnMsg := conn.RemoteAddr().String() + "|SUCCESS"
			messages <- returnMsg
			// Also send a system message to the friend if he/she is online.
			friendAddr := ipnamemap[friend.Username]
			if friendAddr != "" {
				sendMsg := friendAddr + "|SYSTEM|" + user.Username + " has confirmed your friend request."
				messages <- sendMsg
			}

		case "VFL": //Process View Friend List message.
			friends := database.GetFriendList(user, db)
			returnMsg := conn.RemoteAddr().String()
			size := len(friends)
			for i := 0; i < size; i++ {
				status := database.GetUserStatus(friends[i], db)
				returnMsg = returnMsg + "|" + strconv.Itoa(friends[i].Id) + ";" + friends[i].Username + ";" + strconv.Itoa(status)
			}
			messages <- returnMsg

		case "OD": //Process Open Dialog message.
			targetid, _ := strconv.Atoi(strings.Split(recieveStr, "|")[1])
			target := database.GetUserByID(targetid, db)
			chattingTarget[user.Username] = target.Username

			if database.CheckFriendRelationship(target, user, db) == false {
				SendFailMessage(conn, head)
				continue
			}
			returnMsg := conn.RemoteAddr().String() + "|SUCCESS|" + target.Username
			messages <- returnMsg

		case "CD": //Process Close Dialog message.
			chattingTarget[user.Username] = ""

		case "GUM": //Process Get Unread Message message.
			targetName := strings.Split(recieveStr, "|")[1]
			target := database.GetUserByName(targetName, db)

			unreadMessageInfos := database.GetAllUnreadMessage(user, target, rc)

			if unreadMessageInfos.Len() == 0 {
				SendFailMessage(conn, head)
				continue
			}
			sendMsg := conn.RemoteAddr().String()
			for i := 0; unreadMessageInfos.Len() > 0; i++ {
				item := unreadMessageInfos.Front()
				charInfo := item.Value.(*database.ChatInfo)
				encodedBody := charInfo.Body
				sendMsg = sendMsg + "|" + charInfo.Sender + ";" + charInfo.Time + ";" + encodedBody
				unreadMessageInfos.Remove(item)
			}
			messages <- sendMsg

		case "CT": //Process Chat message.
			targetName := strings.Split(recieveStr, "|")[1]
			chatBodyTmp := strings.Split(recieveStr, "|")[2:]
			chatBody := chatBodyTmp[0]
			//To prevent if there are "|"s in body.
			for i := 1; i < len(chatBodyTmp)-1; i++ {
				chatBody = chatBody + "|" + chatBodyTmp[i]
			}

			target := database.GetUserByName(targetName, db)
			timeTmp := time.Now().Unix()
			curTime := time.Unix(timeTmp, 0).String()

			returnMsg := conn.RemoteAddr().String() + "|SUCCESS"
			messages <- returnMsg

			targetAddr := ipnamemap[target.Username]
			database.SaveChatMessage(user, target, curTime, chatBody, db)
			// database.SaveUnreadMessage(user, target, curTime, chatBody, rc)

			if database.CheckStatus(target.Username, db) == false || chattingTarget[target.Username] != user.Username {
				database.SaveUnreadMessage(user, target, curTime, chatBody, rc)
				if database.CheckStatus(target.Username, db) == true {
					returnMsg := targetAddr + "|SYSTEM|You recieve a message from " + user.Username + ". \n"
					messages <- returnMsg
				}
				continue
			}

			if targetAddr != "" {
				sendMsg := targetAddr + "|CHAT|" + user.Username + "|" + curTime + "|" + chatBody
				log.Printf("Send Msg [" + sendMsg + "] from " + user.Username + " to " + target.Username + ". \n")
				messages <- sendMsg
			}

		case "OHD": //Process Open History Dialog message.
			targetid, _ := strconv.Atoi(strings.Split(recieveStr, "|")[1])
			target := database.GetUserByID(targetid, db)

			if database.CheckFriendRelationship(target, user, db) == false {
				SendFailMessage(conn, head)
				continue
			}
			returnMsg := conn.RemoteAddr().String() + "|SUCCESS|" + target.Username
			messages <- returnMsg

			historyChatMsgs = database.GetAllHistory(user, target, db)

		case "VHM": //Process View History Message message.
			if historyChatMsgs.Len() == 0 {
				SendFailMessage(conn, head)
				continue
			}
			sendMsg := conn.RemoteAddr().String()
			for i := 0; i < 10 && historyChatMsgs.Len() > 0; i++ {
				item := historyChatMsgs.Back()
				charInfo := item.Value.(*database.ChatInfo)
				encodedBody := database.Base64Encode(charInfo.Body)
				sendMsg = sendMsg + "|" + charInfo.Sender + ";" + charInfo.Time + ";" + encodedBody
				historyChatMsgs.Remove(item)
			}
			messages <- sendMsg
		default:
			messages <- recieveStr
		}

	}

}

func echoHandler(conns *map[string]net.Conn, messages chan string) {

	for {
		msg := <-messages
		fmt.Println(msg)

		recieverAddr := strings.Split(msg, "|")[0]
		conn := (*conns)[recieverAddr]
		_, err := conn.Write([]byte(msg))
		if err != nil {
			log.Println(err.Error())
			delete(*conns, recieverAddr)
		}
	}

}

func SendFailMessage(conn net.Conn, head string) {
	failMsg := "FAIL|" + head
	log.Printf("Send fail message to ...%s... Fail header is [%s]. \n", conn.RemoteAddr().String(), head)
	_, err := conn.Write([]byte(failMsg))
	if err != nil {
		log.Println(err.Error())
	}
}

func StartServer(port string, db *sql.DB, rc redis.Conn) {
	service := ":" + port //strconv.Itoa(port);
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkError(err, "ResolveTCPAddr")
	l, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err, "ListenTCP")
	conns := make(map[string]net.Conn)
	messages := make(chan string, 10)
	ipnamemap := make(map[string]string)
	chattingTarget := make(map[string]string)

	go echoHandler(&conns, messages)

	for {
		log.Println("Listening ...")
		conn, err := l.Accept()
		checkError(err, "Accept")
		log.Println("Accepting ...")
		conns[conn.RemoteAddr().String()] = conn
		go Handler(conn, messages, ipnamemap, chattingTarget, db, rc)

	}

}

func main() {
	db := database.OpenDatabase()
	rc := database.OpenRedis()
	defer database.CloseRedis(rc)
	defer database.CloseDatabase(db)
	defer database.SetAllStatusOff(db)
	database.FetchOrCreateUserTable(db)
	database.FetchOrCreateFriendTable(db)
	database.FetchOrCreateFriendApplyTable(db)
	database.FetchOrCreateChatInfoTable(db)
	database.FetchOrCreateOfflineChatInfoTable(db)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go SafeExit(c, db, rc)

	StartServer("9090", db, rc)
}

//Safe Exit.
func SafeExit(c chan os.Signal, db *sql.DB, rc redis.Conn) {
	<-c
	fmt.Printf("The server is going to close...\n")
	database.SetAllStatusOff(db)
	fmt.Printf("Set all users' status off...\n")
	database.CloseDatabase(db)
	fmt.Printf("Close the database...\n")
	database.CloseRedis(rc)
	fmt.Printf("Close the Redis...\n")
	fmt.Printf("The server is safely closed.\n")
	os.Exit(0)
}

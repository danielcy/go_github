package main

import (
	"bufio"
	"container/list"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cyang/simpleChat/database"
)

var (
	recieverMutex *sync.Mutex = new(sync.Mutex)
	chatMsg                   = make(chan string)
)

func checkError(err error, info string) (res bool) {

	if err != nil {
		fmt.Println(info + "  " + err.Error())
		return false
	}
	return true
}

func StartClient(port string) {
	tcpaddr := ":" + port
	tcpAddr, err := net.ResolveTCPAddr("tcp4", tcpaddr)
	checkError(err, "ResolveTCPAddr")
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if checkError(err, "DialTCP") == false {
		fmt.Printf("Cannot connect to the server. Please try later. \n")
		return
	}
	fmt.Printf("Already connect to the server. \n")
	defer conn.Close()

	message := make(chan string)

	go ClientProcess(message, conn)

	buf := make([]byte, 1024)
	for {
		lenght, err := conn.Read(buf)
		if checkError(err, "Connection") == false {
			conn.Close()
			fmt.Println("Lost connection to the server...Force to exit.")
			return
		}
		recieveStr := string(buf[0:lenght])
		if len(strings.Split(recieveStr, "|")) > 1 && strings.Split(recieveStr, "|")[1] == "SYSTEM" {
			curTime := GetCurrentTime()
			fmt.Println("[" + curTime + "]" + "[System Message]: " + strings.Split(recieveStr, "|")[2])
		} else if len(strings.Split(recieveStr, "|")) > 1 && strings.Split(recieveStr, "|")[1] == "CHAT" {
			chatMsg <- recieveStr
		} else {
			message <- recieveStr
		}
	}
}

func ClientProcess(message chan string, conn net.Conn) {
	fmt.Printf("Welcome to Simple Chat!\n")
	fmt.Printf("You can print:\n")
	fmt.Printf("	1 to log in;\n")
	fmt.Printf("	2 to sign up;\n")

	choice := ""
	fmt.Scanln(&choice)

	for choice != "1" && choice != "2" {
		fmt.Printf("Please input the right text!\n")
		fmt.Scan(&choice)
	}

	var currentUser *database.User

	if choice == "2" {
		currentUser = SignUp(message, conn)
	} else if choice == "1" {
		currentUser = LogIn(message, conn)
	}

	fmt.Printf("Welcome to Simple Chat, %s!\n", currentUser.Username)
	chatReciever := make(map[string]*list.List)
	chatReminder := make(map[string]*string)
	chatReminder["currentTarget"] = new(string)
	*chatReminder["currentTarget"] = ""
	go RecieveChatInfo(chatReciever, chatReminder, conn)
	for {
		fmt.Printf("Main Menu\n")
		fmt.Printf("You can print:\n")
		fmt.Printf("	1 to add friends;\n")
		fmt.Printf("	2 to chat with friends;\n")
		fmt.Printf("	3 to view pending friend applications;\n")
		fmt.Printf("	4 to view history messages;\n")
		fmt.Printf("	5 to exit;\n")
		SendMessage("CUM", "", "", "", conn)

		choice := ""
		fmt.Scanln(&choice)
		for choice != "1" && choice != "2" && choice != "3" && choice != "4" && choice != "5" {
			fmt.Printf("Please input the right text!\n")
			fmt.Scan(&choice)
		}

		if choice == "1" {
			AddFriend(message, currentUser, conn)
		} else if choice == "2" {
			ChatWithFriends(message, currentUser, chatReciever, chatReminder, conn)
		} else if choice == "3" {
			ConfirmPendingFriendApplication(message, currentUser, conn)
		} else if choice == "4" {
			MainHistoryInterface(message, currentUser, conn)
		} else if choice == "5" {
			fmt.Printf("Goodbye, %s! Thanks for using!\n", currentUser.Username)
			os.Exit(0)
		}
	}

}

func SignUp(message chan string, conn net.Conn) *database.User {
	fmt.Printf("Now you are going to sign up, please enter your: \n")
	for {
		fmt.Printf("Username: ")
		var username string
		fmt.Scanln(&username)
		for len([]rune(username)) < 3 || len([]rune(username)) > 16 {
			fmt.Printf("Username's length should between 3-16. Please try another one. \n")
			fmt.Printf("Username: ")
			fmt.Scanln(&username)
		}

		fmt.Printf("Password: ")
		var password string
		fmt.Scanln(&password)
		for len([]rune(password)) < 6 || len([]rune(password)) > 16 {
			fmt.Printf("Password's length should between 6-16. Please try another one. \n")
			fmt.Printf("Password: ")
			fmt.Scanln(&password)
		}

		SendMessage("SU", username, password, "", conn)
		signUpMsg := <-message

		if strings.Split(signUpMsg, "|")[0] == "FAIL" {
			fmt.Printf("This username exists. Please try another one. \n")
			continue
		}

		userid, _ := strconv.Atoi(strings.Split(signUpMsg, "|")[2])
		var user *database.User = new(database.User)
		user.Id = userid
		user.Username = username
		user.Password = password

		return user
	}
}

func LogIn(message chan string, conn net.Conn) *database.User {
	fmt.Printf("Now you are going to log in, please enter your(If you do not have an account, please enter '/signup'): \n")
	for {
		fmt.Printf("Username: ")
		var username string
		fmt.Scanln(&username)
		if username == "/signup" {
			fmt.Printf("go to sign up. \n")
			result := SignUp(message, conn)
			return result
		}
		fmt.Printf("Password: ")
		var password string
		fmt.Scanln(&password)

		SendMessage("LI", username, password, "", conn)
		logInMsg := <-message

		if strings.Split(logInMsg, "|")[0] == "FAIL" {
			fmt.Printf("Failed to log in. Maybe: \n")
			fmt.Printf("	1.Uncorrect username or password. \n")
			fmt.Printf("	2.This user is already online. \n")
			continue
		}
		userid, _ := strconv.Atoi(strings.Split(logInMsg, "|")[2])
		var user *database.User = new(database.User)
		user.Id = userid
		user.Username = username
		user.Password = password

		return user
	}
}

func AddFriend(message chan string, user *database.User, conn net.Conn) {
	fmt.Printf("You are going to add friends, please enter his/her username.(Enter '/back' to go back to the main menu.) \n")
	for {
		fmt.Printf("Friend Name: ")
		var friendname string
		fmt.Scanln(&friendname)

		if friendname == "/back" {
			return
		}

		SendMessage("AF", friendname, "", "", conn)
		addFriendMsg := <-message
		if strings.Split(addFriendMsg, "|")[0] == "FAIL" {
			fmt.Printf("Failed to send the request. Maybe:  \n")
			fmt.Printf("	1. No such a user.  \n")
			fmt.Printf("	2. You are already friends.  \n")
			fmt.Printf("	3. You or he/she have already send a request.  \n")
			continue
		}
		fmt.Printf("You've sent a friend request to %s. \n", friendname)
	}
}

func ConfirmPendingFriendApplication(message chan string, user *database.User, conn net.Conn) {
	fmt.Printf("The following are the pending friend applications: \n")
	SendMessage("VRL", "", "", "", conn)
	requestListMsg := <-message
	requestList := strings.Split(requestListMsg, "|")[1:]
	for i := 0; i < len(requestList); i++ {
		request := requestList[i]
		requestID := strings.Split(request, ";")[0]
		requestName := strings.Split(request, ";")[1]
		fmt.Printf("UserID: %s 		Username: %s \n", requestID, requestName)
	}
	fmt.Printf("Please enter the friend's id to apply his/her application.(Enter '/back' to go back to the main menu.) \n")
	var input string
	for {
		fmt.Printf("ID: ")
		fmt.Scanln(&input)
		if input == "/back" {
			return
		}
		_, err := strconv.Atoi(input)
		if err != nil {
			fmt.Printf("Invalid Input. Please enter an ID number. \n")
			continue
		}
		SendMessage("CR", input, "", "", conn)
		confirmRequestMsg := <-message

		if strings.Split(confirmRequestMsg, "|")[0] == "FAIL" {
			fmt.Printf("Failed to add this person. Maybe he/she is not in your list. :  \n")
			continue
		}
		fmt.Printf("Confirm operation success. \n")

	}
}

func ChatWithFriends(message chan string, user *database.User, chatReciever map[string]*list.List, chatReminder map[string]*string, conn net.Conn) {
	fmt.Printf("Friends List: \n")
	SendMessage("VFL", "", "", "", conn)
	friendListMsg := <-message
	friendList := strings.Split(friendListMsg, "|")[1:]
	for i := 0; i < len(friendList); i++ {
		friend := friendList[i]
		friendID := strings.Split(friend, ";")[0]
		friendName := strings.Split(friend, ";")[1]
		stat := strings.Split(friend, ";")[2]
		var friendStatus string
		if stat == "0" {
			friendStatus = "Offline"
		} else {
			friendStatus = "Online"
		}
		fmt.Printf("UserID: %s 	Username: %s 	Status: %s\n", friendID, friendName, friendStatus)
	}
	fmt.Printf("Please enter friend's id to chat with him.(Enter '/back' to go back to the main menu.) \n")
	for {
		fmt.Printf("ID: ")
		var input string
		fmt.Scanln(&input)
		if input == "/back" {
			return
		}
		_, err := strconv.Atoi(input)
		if err != nil {
			fmt.Printf("Invalid Input. Please try again. \n")
			continue
		}
		SendMessage("OD", input, "", "", conn)
		openDialogMsg := <-message
		if strings.Split(openDialogMsg, "|")[0] == "FAIL" {
			fmt.Printf("This ID is not available. Please ensure the user is your friend and is online.  \n")
			continue
		}
		targetName := strings.Split(openDialogMsg, "|")[2]
		target := new(database.User)
		target.Id, _ = strconv.Atoi(input)
		target.Username = targetName
		Chat(message, user, target, chatReciever, chatReminder, conn)
	}

}

func Chat(message chan string, user *database.User, target *database.User, chatReciever map[string]*list.List, chatReminder map[string]*string, conn net.Conn) {
	fmt.Printf("Now you are chatting with %s. Enter '/back' to return to friend list. \n \n", target.Username)
	*chatReminder["currentTarget"] = target.Username
	signal := make(map[string]*string)
	signal["kill"] = new(string)
	GetUnreadMessageFromTarget(message, user, target, conn)
	go GetChatInfoAndPrint(chatReciever, target, signal)
	//go GetChatInfoAndPrint_(target, signal, conn)
	for {
		inputReader := bufio.NewReader(os.Stdin)
		var chatbody string
		input, err := inputReader.ReadString('\n')
		if err == nil {
			chatbody = input
		}
		if chatbody == "/back\n" {
			*signal["kill"] = "1"
			*chatReminder["currentTarget"] = ""
			SendMessage("CD", "", "", "", conn)
			return
		}
		if chatbody == "\n" {
			continue
		}
		curTime := GetCurrentTime()
		fmt.Printf("\n[%s][Your Word]: %s\n", curTime, chatbody)
		SendMessage("CT", target.Username, chatbody, "", conn)
		returnMsg := <-message
		if strings.Split(returnMsg, "|")[0] == "FAIL" {
			fmt.Printf("Failed to send the message. %s is not online. You can chat with other friends. \n", target.Username)
			return
		}
	}
}

func MainHistoryInterface(message chan string, user *database.User, conn net.Conn) {
	fmt.Printf("Friends List: \n")
	SendMessage("VFL", "", "", "", conn)
	friendListMsg := <-message
	friendList := strings.Split(friendListMsg, "|")[1:]
	for i := 0; i < len(friendList); i++ {
		friend := friendList[i]
		friendID := strings.Split(friend, ";")[0]
		friendName := strings.Split(friend, ";")[1]
		fmt.Printf("UserID: %s 	Username: %s \n", friendID, friendName)
	}
	fmt.Printf("Please enter friend's id to view history message.(Enter '/back' to go back to the main menu.) \n")
	for {
		fmt.Printf("ID: ")
		var input string
		fmt.Scanln(&input)
		if input == "/back" {
			return
		}
		_, err := strconv.Atoi(input)
		if err != nil {
			fmt.Printf("Invalid Input. Please try again. \n")
			continue
		}
		SendMessage("OHD", input, "", "", conn)
		returnMsg := <-message
		if strings.Split(returnMsg, "|")[0] == "FAIL" {
			fmt.Printf("This ID is not available. Please ensure the user is your friend.  \n")
			continue
		}
		targetName := strings.Split(returnMsg, "|")[2]
		target := new(database.User)
		target.Id, _ = strconv.Atoi(input)
		target.Username = targetName
		ViewHistoryMessage(message, user, target, conn)
	}

}

func ViewHistoryMessage(message chan string, user *database.User, target *database.User, conn net.Conn) {
	fmt.Printf("History message with %s: \n", target.Username)
	for {
		fmt.Printf("Enter 'n' to get next 10 messages; Enter '/back' to return. \n")
		input := ""
		fmt.Scanln(&input)
		if input == "/back" {
			return
		}
		if input != "n" {
			fmt.Printf("Invalid input. \n")
			continue
		}
		SendMessage("VHM", "", "", "", conn)
		returnMsg := <-message
		if strings.Split(returnMsg, "|")[0] == "FAIL" {
			fmt.Printf("No more history log. \n")
			continue
		}
		chatMsg := strings.Split(returnMsg, "|")[1:]
		for i := 0; i < len(chatMsg); i++ {
			splitInfos := strings.Split(chatMsg[i], ";")
			senderName := splitInfos[0]
			curTime := splitInfos[1]
			encodeBody := splitInfos[2]
			body, _ := database.Base64Decode(encodeBody)
			fmt.Printf("[%s][%s]: %s\n", curTime, senderName, body)
		}
	}
}

func SendMessage(head string, str1 string, str2 string, str3 string, conn net.Conn) {
	message := head + "|" + str1 + "|" + str2 + "|" + str3
	_, err := conn.Write([]byte(message))
	if err != nil {
		fmt.Println(err.Error())
		conn.Close()
		return
	}
}

func RecieveChatInfo(chatReciever map[string]*list.List, chatReminder map[string]*string, conn net.Conn) {
	for {
		msg := <-chatMsg
		recieverMutex.Lock()
		// fmt.Printf("[DEBUG][RECIEVE CHAT MESSAGE]: %s \n", msg)
		senderName := strings.Split(msg, "|")[2]

		if chatReciever[senderName] == nil {
			chatReciever[senderName] = list.New()
		}
		chatReciever[senderName].PushBack(msg)
		/*if *chatReminder["currentTarget"] != senderName {
			curTime := GetCurrentTime()
			fmt.Printf("[%s][System Message]: You recieve a message from %s. \n", curTime, senderName)
		}*/
		recieverMutex.Unlock()
		//fmt.Printf("[LOG]Reading unlocked.")
	}
}

func GetChatInfoAndPrint(chatReciever map[string]*list.List, sender *database.User, signal map[string]*string) {
	for {
		recieverMutex.Lock()
		recieverMutex.Unlock()
		chatMsgList := chatReciever[sender.Username]
		if chatMsgList == nil || chatMsgList.Len() == 0 {
			continue
		}
		if *signal["kill"] == "1" {
			return
		}
		firstChatMsgItem := chatMsgList.Front()
		chatMsgList.Remove(firstChatMsgItem)
		firstChatMsg := firstChatMsgItem.Value.(string)

		senderName := strings.Split(firstChatMsg, "|")[2]
		time := strings.Split(firstChatMsg, "|")[3]
		bodyTmp := strings.Split(firstChatMsg, "|")[4:]
		body := bodyTmp[0]
		//To prevent if there are "|"s in body.
		for i := 1; i < len(bodyTmp); i++ {
			body = body + "|" + bodyTmp[i]
		}

		fmt.Printf("[%s][%s]: %s\n", time, senderName, body)
	}
}

func GetUnreadMessageFromTarget(message chan string, user *database.User, target *database.User, conn net.Conn) {
	SendMessage("GUM", target.Username, "", "", conn)
	returnMsg := <-message

	if strings.Split(returnMsg, "|")[0] == "FAIL" {
		return
	}

	unreadMsgs := strings.Split(returnMsg, "|")[1:]
	for i := 0; i < len(unreadMsgs); i++ {
		splitInfos := strings.Split(unreadMsgs[i], ";")
		senderName := splitInfos[0]
		curTime := splitInfos[1]
		encodeBody := splitInfos[2]
		body, _ := database.Base64Decode(encodeBody)
		fmt.Printf("[%s][%s]: %s\n", curTime, senderName, body)
	}

}

func GetCurrentTime() string {
	timeTmp := time.Now().Unix()
	curTime := time.Unix(timeTmp, 0).String()
	return curTime
}

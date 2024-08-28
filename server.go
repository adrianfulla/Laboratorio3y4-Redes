package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/adrianfulla/Proyecto1-Redes/server/xmpp-functions"
)


// func main() {
// 	LogInTest()
// }

func LogInTest(){
	conn, err := xmppfunctions.Login("alumchat.lol", "5222", "aa-test3", "12345")
	if err != nil{
		fmt.Println(err.Error())
		return 
	}

	// xmppfunctions.AddContact(conn, "afp21592@alumchat.lol")

	// xmppfunctions.SendMessage(conn, "afp21592@alumchat.lol", "Hello there!")

	// xmppfunctions.GetContacts(conn)
	conn.SendPresence("presence", "Online")
	xmppfunctions.ReceiveMessages(conn)
}

func CreateUserTest(){
	err := xmppfunctions.CreateUser("alumchat.lol", "5222", "aa-test3", "12345")
	if err != nil{
		fmt.Println(err.Error())
		return 
	}
}

func homePageHandler(w http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprintf(w, "hello world")
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		log.Panic(err)
	}
}

type thumbnailRequest struct {
	Url string `json:"url"`
}

func thumbnailHandler(w http.ResponseWriter, r *http.Request) {
	var decoded thumbnailRequest

	err := json.NewDecoder(r.Body).Decode(&decoded)
	if err != nil {
		fmt.Printf("Got a thumbnail")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Printf("Got the following url: %s\n", decoded.Url)
}

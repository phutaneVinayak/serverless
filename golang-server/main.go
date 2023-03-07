package main

import (
	// "errors"
	"encoding/json"
	"fmt"
	"time"

	// "io/ioutil"
	"net/http"

	"github.com/gorilla/mux"

	pubnub "github.com/pubnub/go/v7"
	// "os"
)

type Movie struct {
	Id    string `json:"id"`
	Isbn  string `json:"isbn"`
	Title string `json:"title"`
}

type Payload struct {
	FunctionName string `json:"functionName"`
	UserData     string `json:"userData"`
}

var movieOne = Movie{
	Id:    "Hello",
	Isbn:  "bla bla",
	Title: "bla bla",
}

var msg = map[string]interface{}{
	"msg": "Hello worlds",
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Println("This is server")
	// http.Error(w, "÷unable to process request", 400)
	json.NewEncoder(w).Encode(movieOne)
}

func wrapperHandler(pn *pubnub.PubNub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var requestPayload Payload
		_ = json.NewDecoder(r.Body).Decode(&requestPayload)
		fmt.Println("This is value from user %v", requestPayload)

		// pubnub for listener
		listener := pubnub.NewListener()
		// doneConnect := make(chan bool)
		// donePublish := make(chan bool)
		doneReceviced := make(chan bool)

		pn.Subscribe().
			Channels([]string{"hello_world"}).
			Execute()

		// var msgResponse map[string]interface{}
		go func() {
			fmt.Println("inside 1")
			for {
				fmt.Println("inside for 1")
				select {
				case message := <-listener.Message:
					fmt.Println("inside case 1")
					// Handle new message stored in message.message
					if msg, ok := message.Message.(map[string]interface{}); ok {
						fmt.Println("inside if condition 1")
						fmt.Println("erewrwe", msg["msg"])

						json.NewEncoder(w).Encode(msg)

					}
					doneReceviced <- true
				}
			}
		}()

		fmt.Println("outside after go func ")
		pn.AddListener(listener)

		fmt.Println("Before doneRececi check")
		<-doneReceviced

		// fmt.Println("msgResponse", msgResponse)
	}
}

func getExecution(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var requestPayload Payload
	_ = json.NewDecoder(r.Body).Decode(&requestPayload)
	fmt.Println("This is value from user %v", requestPayload)
}

func main() {
	config := pubnub.NewConfigWithUserId("myUniqueUserId")
	config.SubscribeKey = "x-x-x-x-x"
	config.PublishKey = "x-x-x-x-x-x-x-x"

	pn := pubnub.NewPubNub(config)

	time.AfterFunc(10*time.Second, func() {
		response, status, err1 := pn.Publish().
			Channel("hello_world").Message(msg).Execute()

		if err1 != nil {
			// Request processing failed.
			// Handle message publish error
		}

		fmt.Println(response, status, err1)
	})

	routes := mux.NewRouter()

	routes.HandleFunc("/", getRoot)
	routes.HandleFunc("/execute", wrapperHandler(pn))
	err := http.ListenAndServe(":3333", routes)

	fmt.Println("Error while starting server %v", err)
}

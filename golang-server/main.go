package main

import (
	// "errors"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	// "io/ioutil"
	"net/http"

	"github.com/gorilla/mux"

	pubnub "github.com/pubnub/go/v7"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/snappy"
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

var writer *kafka.Writer

func Configure(kafkaBrokerUrls []string, clientId string, topic string) (w *kafka.Writer, err error) {
	dialer := &kafka.Dialer{
		Timeout:  10 * time.Second,
		ClientID: clientId,
	}

	config := kafka.WriterConfig{
		Brokers:          kafkaBrokerUrls,
		Topic:            topic,
		Balancer:         &kafka.LeastBytes{},
		Dialer:           dialer,
		WriteTimeout:     10 * time.Second,
		ReadTimeout:      10 * time.Second,
		CompressionCodec: snappy.NewCompressionCodec(),
	}
	w = kafka.NewWriter(config)
	writer = w
	return w, nil
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

		// send event and wait for response
		/**
		This function will create one unique channel ID and pass it to sendJobEvent function, than this channel id will send to listener to wait for event
		*/
		sendJobEvent("12344", requestPayload)

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
	}
}

func getExecution(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var requestPayload Payload
	_ = json.NewDecoder(r.Body).Decode(&requestPayload)
	fmt.Println("This is value from user %v", requestPayload)
}

func sendJobEvent(channelEvent string, payload Payload) {
	parent := context.Background()
	defer parent.Done()

	formInBytes, err := json.Marshal(payload)

	if err != nil {
		fmt.Sprintf("error while marshalling json: %s", err.Error())
		return
	}

	message := kafka.Message{
		Key:   nil,
		Value: formInBytes,
		Time:  time.Now(),
	}

	err = writer.WriteMessages(parent, message)

	if err != nil {
		fmt.Sprintf("error while push message into kafka: %s", err.Error())
		return
	}

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

	// init kafka
	// strings.Split(kafkaBrokerUrl, ","), kafkaClientId, kafkaTopic
	var kafkaBrokerUrl = "http://kafka-broker-ingress.knative-eventing.svc.cluster.local/knative-eventing/kafka-event-broker"
	kafkaProducer, err2 := Configure(strings.Split(kafkaBrokerUrl, ","), "kafkaClientId", "kafkaTopic")
	if err2 != nil {
		fmt.Println("unable to configure kafka")
		return
	}
	defer kafkaProducer.Close()

	routes := mux.NewRouter()

	routes.HandleFunc("/", getRoot)
	routes.HandleFunc("/execute", wrapperHandler(pn))
	err := http.ListenAndServe(":3333", routes)

	fmt.Println("Error while starting server %v", err)
}

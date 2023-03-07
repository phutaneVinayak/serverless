listener1 := pubnub.NewListener()
doneConnect := make(chan bool)
donePublish := make(chan bool)

go func() {
	for {
		select {
		case status := <-listener1.Status:
			switch status.Category {
			case pubnub.PNDisconnectedCategory:
				// This event happens when radio / connectivity is lost
			case pubnub.PNConnectedCategory:
				// Connect event. You can do stuff like publish, and know you'll get it.
				// Or just use the connected event to confirm you are subscribed for
				// UI / internal notifications, etc
				doneConnect <- true
			case pubnub.PNReconnectedCategory:
				// Happens as part of our regular operation. This event happens when
				// radio / connectivity is lost, then regained.
			}
		case message := <-listener1.Message:
			// Handle new message stored in message.message
			if message.Channel != "" {
				// Message has been received on channel group stored in
				// message.Channel
			} else {
				// Message has been received on channel stored in
				// message.Subscription
			}
			if msg, ok := message.Message.(map[string]interface{}); ok {
				fmt.Println("erewrwe", msg["msg"])
			}
			/*
			   log the following items with your favorite logger
			       - message.Message
			       - message.Subscription
			       - message.Timetoken
			*/

			donePublish <- true
		case <-listener1.Presence:
			// handle presence
		}
	}
}()

pn.AddListener(listener1)

pn.Subscribe().
	Channels([]string{"hello_world"}).
	Execute()

<-doneConnect

response, status, err1 := pn.Publish().
	Channel("hello_world").Message(msg).Execute()

if err1 != nil {
	// Request processing failed.
	// Handle message publish error
}

fmt.Println(response, status, err1)

<-donePublish
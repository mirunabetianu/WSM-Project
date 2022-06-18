package utils

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var mqttBroker = "localhost"
var mqttPort = 1883
var mqttClientId = "order_service_id"
var mqttUsername = "order_service"
var mqttPassword = "public"

var TOPIC_ADD_ITEM = "topic/addItem"
var TOPIC_REMOVE_ITEM = "topic/removeItem"

//var Chans = make(map[string]chan int)

var ItemChannel = make(chan string)

func OpenMqttConnection() mqtt.Client {
	// init required options
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", mqttBroker, mqttPort))
	opts.SetClientID(mqttClientId)
	opts.SetUsername(mqttUsername)
	opts.SetPassword(mqttPassword)
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	// create the client with options and connect
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	return client
}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	go func(chan string) {
		println(string(msg.Payload()))
		switch {
		case msg.Topic() == "topic/addItemResponse":
			var itemId, itemPrice, status, orderId int
			_, err := fmt.Sscanf(string(msg.Payload()), "orderId:%d-itemId:%d-price:%d-status:%d", &orderId, &itemId, &itemPrice, &status)
			channelKey := fmt.Sprintf("orderId:%d-itemId:%d-price:%d", orderId, itemId, itemPrice)

			println(status)
			if err != nil || status == 500 {
				ItemChannel <- "error"
				//print(len(Chans))
			} else {
				ItemChannel <- channelKey
				//Chans[channelKey] <- itemPrice
			}
		default:
		}
	}(ItemChannel)
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

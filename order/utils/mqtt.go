package utils

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"math"
)

var mqttBroker = "localhost"
var mqttPort = 1883
var mqttClientId = "order_service_id"
var mqttUsername = "order_service"
var mqttPassword = "public"

type ItemChannel struct {
	OrderId int
	ItemId  int
	Channel chan int
}

var Chans []ItemChannel

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
	switch {
	case msg.Topic() == "topic/findItemResponse":
		var itemId, itemPrice, status, orderId int
		_, err := fmt.Sscanf(string(msg.Payload()), "orderId:%d-itemId:%d-price:%d-status:%d", &orderId, &itemId, &itemPrice, &status)
		var index int
		for i := range Chans {
			x := len(Chans) - i - 1
			if Chans[x].OrderId == orderId && Chans[x].ItemId == itemId {
				index = x
				break
			}
		}
		go func(chan int) {
			if err != nil || status == 500 {
				Chans[index].Channel <- math.MaxInt
			} else {
				Chans[index].Channel <- itemPrice
			}
		}(Chans[index].Channel)
	default:
	}

}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

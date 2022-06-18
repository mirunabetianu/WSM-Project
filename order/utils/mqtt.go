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

type CheckoutItem struct {
	OrderId        int
	PaymentChannel chan string
	StockChannel   chan string
}

var ItemChannels []ItemChannel
var CheckoutChannels []CheckoutItem

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
		for i := range ItemChannels {
			x := len(ItemChannels) - i - 1
			if ItemChannels[x].OrderId == orderId && ItemChannels[x].ItemId == itemId {
				index = x
				break
			}
		}
		go func(chan int) {
			if err != nil || status == 500 {
				ItemChannels[index].Channel <- math.MaxInt
			} else {
				ItemChannels[index].Channel <- itemPrice
			}
		}(ItemChannels[index].Channel)
	case msg.Topic() == "topic/subtractStockResponse":
		var index int
		var payload string
		var orderId int
		_, err := fmt.Sscanf(string(msg.Payload()), "orderId:%d-%s", &orderId, &payload)

		for i := range CheckoutChannels {
			x := len(CheckoutChannels) - i - 1
			if CheckoutChannels[x].OrderId == orderId {
				index = x
				break
			}
		}
		go func(chan string) {
			if payload == "error" || err != nil {
				CheckoutChannels[index].StockChannel <- "error"
			} else {
				CheckoutChannels[index].StockChannel <- "success"
			}
		}(CheckoutChannels[index].StockChannel)
	case msg.Topic() == "topic/paymentResponse":
		var index int
		var payload string
		var orderId int
		_, err := fmt.Sscanf(string(msg.Payload()), "orderId:%d-%s", &orderId, &payload)
		for i := range CheckoutChannels {
			x := len(CheckoutChannels) - i - 1
			if CheckoutChannels[x].OrderId == orderId {
				index = x
				break
			}
		}
		go func(chan string) {
			if payload == "error" || err != nil {
				CheckoutChannels[index].PaymentChannel <- "error"
			} else {
				CheckoutChannels[index].PaymentChannel <- "success"
			}
		}(CheckoutChannels[index].PaymentChannel)
	default:
	}

}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

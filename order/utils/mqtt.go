package utils

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"math"
)

var id = uuid.New()
var mqttBroker = "localhost"
var mqttPort = 1883
var mqttClientId = "order_service_id" + id.String()
var mqttUsername = "order_service" + id.String()
var mqttPassword = "public"

type ItemChannel struct {
	Id      string
	OrderId int
	ItemId  int
	Channel chan int
}

type CheckoutItem struct {
	Id             uint32
	OrderId        int
	PaymentChannel chan string
	StockChannel   chan string
}

type RefundItem struct {
	Id            uint32
	RefundChannel chan string
}

var ItemChannels []ItemChannel
var CheckoutChannels []CheckoutItem
var RefundChannels []RefundItem

func OpenMqttConnection() mqtt.Client {
	if GetEnv("EMQX_BROKER_SERVICE_HOST") != "" {
		mqttBroker = GetEnv("EMQX_BROKER_SERVICE_HOST")
	}
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
		var id string
		_, err := fmt.Sscanf(string(msg.Payload()), "orderId:%d-itemId:%d-price:%d-status:%d-id:%s", &orderId, &itemId, &itemPrice, &status, &id)
		var index int
		for i := range ItemChannels {
			x := len(ItemChannels) - i - 1
			if ItemChannels[x].Id == id {
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
		var id uint32
		_, err := fmt.Sscanf(string(msg.Payload()), "orderId:%d-id:%d-%s", &orderId, &id, &payload)

		for i := range CheckoutChannels {
			x := len(CheckoutChannels) - i - 1
			if CheckoutChannels[x].Id == id {
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
		var id uint32
		_, err := fmt.Sscanf(string(msg.Payload()), "orderId:%d-id:%d-%s", &orderId, &id, &payload)
		for i := range CheckoutChannels {
			x := len(CheckoutChannels) - i - 1
			if CheckoutChannels[x].Id == id {
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
	case msg.Topic() == "topic/refundResponse":
		var index int
		var payload string
		var id uint32
		_, err := fmt.Sscanf(string(msg.Payload()), "id:%d-%s", &id, &payload)
		for i := range RefundChannels {
			x := len(RefundChannels) - i - 1
			if RefundChannels[x].Id == id {
				index = x
				break
			}
		}
		go func(chan string) {
			if payload == "error" || err != nil {
				RefundChannels[index].RefundChannel <- "error"
			} else {
				RefundChannels[index].RefundChannel <- "success"
			}
		}(RefundChannels[index].RefundChannel)
	default:
	}

}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

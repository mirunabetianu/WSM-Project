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

func OpenMqttConnection() mqtt.Client {
	// init required options
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", mqttBroker, mqttPort))
	opts.SetClientID(mqttClientId)
	opts.SetUsername(mqttUsername)
	opts.SetPassword(mqttPassword)
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.SetOrderMatters(false)
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
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

func Publish(payload interface{}, client mqtt.Client, topic string) {
	fmt.Printf("\npublishing to topic: %v with payload: %v\n", topic, payload)
	token := client.Publish(topic, 0, true, payload)
	token.Wait()
	//time.Sleep(time.Second)
}

func Subscribe(client mqtt.Client, topic string) {
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
	fmt.Printf("subscribing to topic: %v", topic)
}

func SubscribeForResponse(client mqtt.Client, topic string, responseReceived mqtt.MessageHandler) {
	token := client.Subscribe(topic, 1, responseReceived)
	token.Wait()
	fmt.Printf("subscribing to (response) topic: %v", topic)
}

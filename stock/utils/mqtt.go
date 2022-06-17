package utils

import (
	"fmt"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var mqttBroker = "localhost"
var mqttPort = 1883
var mqttClientId = "stock_service_id"
var mqttUsername = "stock_service"
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

//@input func function --> the function that the mqtt callback will call
//@output func: a mqtt callback function
var generateMessageHandler = func(function any) mqtt.MessageHandler {
	var callBack mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
		payload := strings.Split(string(msg.Payload()), "/")
		var f = function.(func([]string) (string, string))
		if !msg.Retained() {
			res, err := f(payload)
			response_id := payload[len(payload)-1]
			Publish(res+"/"+err+"/"+response_id, client, "topic/response")
		}
	}
	return callBack
}

var hadnleSubtract mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
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
}

func Subscribe(client mqtt.Client, topic string, fun func([]string) (string, string)) {
	fmt.Printf("subscribing to topic: %v", topic)
	var handler_function = generateMessageHandler(fun)
	token := client.Subscribe(topic, 0, handler_function)
	token.Wait()
}

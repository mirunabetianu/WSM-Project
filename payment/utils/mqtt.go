package utils

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

var id = uuid.New()
var mqttBroker = "localhost"
var mqttPort = 1883
var mqttClientId = "payment_service_id" + id.String()
var mqttUsername = "payment_service" + id.String()
var mqttPassword = "public"

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

}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("OpenPsqlConnection lost: %v", err)
}

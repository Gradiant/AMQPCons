package main

import (
	"encoding/json"
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func connectMQTT() (mqtt.Client, error) {

	opts := mqtt.NewClientOptions().AddBroker("tcp://" + *mqttBroker)

	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("Error connecting to MQTT broker: %s", token.Error())
	}

	return client, nil
}

//Callback called when a message arrives on the subscribed topic
func mqttCallback(client mqtt.Client, msg mqtt.Message) {

	log.Printf("Configuring consumer")

	var jsonMsg Message
	err := json.Unmarshal(msg.Payload(), &jsonMsg)

	if err != nil {
		log.Printf("Error parsing JSON: %s", err)
	}

	switch jsonMsg.Type {
	case "NewConsumer":
		newConsumer(jsonMsg.Exchange, jsonMsg.ExchangeType, jsonMsg.BindingKey, jsonMsg.Queue, jsonMsg.DeclareQueue, rabbitConn, client)
	case "DelConsumer":
		deleteConsumer(jsonMsg.Ctag, client)
	}

}

func publishMqtt(topic string, message string, client mqtt.Client) error {

	if token := client.Publish(topic, 0, false, message); token.Wait() && token.Error() != nil {
		return fmt.Errorf("Error publishing to topic %s: %s", topic, token.Error())
	}

	return nil
}

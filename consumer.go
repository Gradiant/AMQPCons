package main

import (
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/satori/go.uuid"
	"github.com/streadway/amqp"
)

func createExchange(exchange, exchangeType string, channel *amqp.Channel) error {

	log.Printf("Declaring Exchange (%q)", exchange)
	if err := channel.ExchangeDeclare(
		exchange,     // name of the exchange
		exchangeType, // type
		true,         // durable
		false,        // delete when complete
		false,        // internal
		false,        // noWait
		nil,          // arguments
	); err != nil {
		return fmt.Errorf("Exchange Declare: %s", err)
	}

	return nil

}

func createQueue(queueName string, channel *amqp.Channel) error {

	log.Printf("Declaring Queue %q", queueName)
	_, err := channel.QueueDeclare(
		queueName, // name of the queue
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // noWait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("Queue Declare: %s", err)
	}

	return nil
}

func bindQueue(queueName, exchange, key string, channel *amqp.Channel) error {

	log.Printf("Binding to Exchange (key %q)", key)

	if err := channel.QueueBind(
		queueName, // name of the queue
		key,       // bindingKey
		exchange,  // sourceExchange
		false,     // noWait
		nil,       // arguments
	); err != nil {
		return fmt.Errorf("Queue Bind: %s", err)
	}

	return nil
}

func newConsumer(exchange, exchangeType, key, queueName string, declareQueue bool, connection *amqp.Connection, mqttClient mqtt.Client) error {

	channel, err := connection.Channel()
	if err != nil {
		return fmt.Errorf("Error getting channel: %s", err)
	}

	if exchange != "" && exchangeType != "" {
		if err = createExchange(exchange, exchangeType, channel); err != nil {
			return err
		}
	}

	if queueName != "" && declareQueue {
		if err = createQueue(queueName, channel); err != nil {
			return err
		}
	}

	if key != "" {
		if err = bindQueue(queueName, exchange, key, channel); err != nil {
			return err
		}
	}

	ctag, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("Generating ctag: %s", err)
	}

	log.Printf("Consumer tag %q", ctag)

	deliveries, err := channel.Consume(
		queueName,     // name
		ctag.String(), // consumerTag,
		false,         // noAck
		false,         // exclusive
		false,         // noLocal
		false,         // noWait
		nil,           // arguments
	)
	if err != nil {
		return fmt.Errorf("Queue Consume: %s", err)
	}

	consumer := &Consumer{
		Exchange:     exchange,
		ExchangeType: exchangeType,
		Queue:        queueName,
		BindingKey:   key,
		Channel:      channel,
		ErrorCh:      make(chan error),
	}

	consumerMap[ctag.String()] = *consumer

	go handleDeliveries(ctag.String(), deliveries, consumer.ErrorCh, mqttClient)

	publishMqtt(*subQueueTopic+"/ctags", "{'"+ctag.String()+"'}", mqttClient)

	return nil

}

func handleDeliveries(ctag string, deliveries <-chan amqp.Delivery, done chan error, mqttClient mqtt.Client) {

	log.Printf("Waiting for messages")

	for d := range deliveries {
		log.Printf(
			"got %dB delivery: [%v] %q",
			len(d.Body),
			d.DeliveryTag,
			d.Body,
		)
		publishMqtt(*pubDataTopic, "{'ctag':'"+ctag+"','payload':'"+string(d.Body)+"'}", mqttClient)
		d.Ack(false)
	}
	log.Printf("handle: deliveries channel closed")
	done <- nil

}

func deleteConsumer(ctag string, mqttClient mqtt.Client) {

	consumer := consumerMap[ctag]
	consumer.Channel.Close()
	delete(consumerMap, ctag)

}

package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/streadway/amqp"
)

func connectAMQP(client mqtt.Client) *amqp.Connection {

	for {

		log.Printf("Dialing without TLS: %s", *uri)

		rabbitConn, err := amqp.Dial("amqp://" + *uri)
		if err == nil {
			log.Printf("Connection established. Getting channel")
			connectionIsUp = true
			notifyConnectionStatus(client, connectionIsUp)
			return rabbitConn
		}

		log.Printf("Trying to reconnect...")
		connectionIsUp = false
		notifyConnectionStatus(client, connectionIsUp)
		time.Sleep(5 * time.Second)
	}

}

func connectAMQPWithTLS(client mqtt.Client) *amqp.Connection {

	cfg := new(tls.Config)
	cfg.RootCAs = x509.NewCertPool()

	cfg.InsecureSkipVerify = true

	//Read CA certificate
	if ca, err := ioutil.ReadFile(*cacertpath); err == nil {
		cfg.RootCAs.AppendCertsFromPEM(ca)
	} else if err != nil {
		log.Fatalf("Error reading CA certificate: %s", err)
		os.Exit(1)
	}

	//Read client key and certificate
	if cert, err := tls.LoadX509KeyPair(*clcertpath, *clkeypath); err == nil {
		cfg.Certificates = append(cfg.Certificates, cert)
	} else if err != nil {
		log.Fatalf("Error reading client certificate or key: %s", err)
		os.Exit(1)
	}

	for {
		log.Printf("Dialing with TLS: %s", *uri)

		rabbitConn, err := amqp.DialTLS("amqps://"+*uri, cfg)
		if err == nil {
			log.Printf("Connection established")
			connectionIsUp = true
			notifyConnectionStatus(client, connectionIsUp)
			return rabbitConn
		}

		log.Printf("Trying to reconnect...")
		connectionIsUp = false
		notifyConnectionStatus(client, connectionIsUp)
		time.Sleep(5 * time.Second)

	}

}

func rabbitConnector(client mqtt.Client) {
	var rabbitErr *amqp.Error

	for {
		rabbitErr = <-rabbitCloseError

		if rabbitErr != nil {

			if usingTLS {
				rabbitConn = connectAMQPWithTLS(client)
			} else {
				rabbitConn = connectAMQP(client)
			}

			rabbitCloseError = make(chan *amqp.Error)
			rabbitConn.NotifyClose(rabbitCloseError)
		}
	}
}

func notifyConnectionStatus(client mqtt.Client, connectionIsUp bool) {

	var msg string = "0"
	if connectionIsUp {
		msg = "1"
	}

	if token := client.Publish(*pubConnTopic, 0, false, msg); token.Wait() && token.Error() != nil {
		log.Printf("Error publishing message to MQTT broker: %s", token.Error())
	}

}

package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/streadway/amqp"
)

var (
	uri           = flag.String("uri", "", "AMQP URI (Mandatory). Eg.: guest:guest@localhost:5672/vhost ")
	mqttBroker    = flag.String("mqttBroker", "", "MQTT broker URI (Mandatory). Eg: 192.168.1.1:1883")
	cacertpath    = flag.String("cacert", "", "Path to the Root CA certificate")
	clcertpath    = flag.String("clcert", "", "Path to the client certificate")
	clkeypath     = flag.String("clkey", "", "Path to the client key")
	pubConnTopic  = flag.String("connTopic", "", "MQTT topic used to publish connection status changes (Mandatory). Eg: consumer/conn")
	pubDataTopic  = flag.String("dataTopic", "", "MQTT topic used to subscribe to data other services want to publish (Mandatory). Eg: consumer/data")
	subQueueTopic = flag.String("queueTopic", "", "MQTT topic used to indicate the consumer that we want to receive messages from an AMQP queue (Mandatory). Eg: consumer/queue")
)

var (
	rabbitConn       *amqp.Connection
	rabbitCloseError chan *amqp.Error
	usingTLS         bool
	connectionIsUp   bool
	consumerMap      map[string]Consumer
)

type Message struct {
	Type         string
	Exchange     string
	ExchangeType string
	Queue        string
	DeclareQueue bool
	BindingKey   string
	Ctag         string
}

type Consumer struct {
	Exchange     string
	ExchangeType string
	Queue        string
	BindingKey   string
	Channel      *amqp.Channel
	ErrorCh      chan error
}

func init() {
	flag.Parse()
	consumerMap = make(map[string]Consumer)
}

//******* MAIN ********/
func main() {

	//Check if required arguments have been specified
	if *uri == "" || *mqttBroker == "" || *subQueueTopic == "" || *pubConnTopic == "" || *pubDataTopic == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	//Channel used to block while receiving messages
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	clientMQTT, err := connectMQTT()
	if err != nil {
		log.Fatalf("Error connecting to MQTT broker: %s", err)
	}

	//Check if certificates are being passed to use TLS or not
	if *cacertpath == "" || *clcertpath == "" || *clkeypath == "" {
		usingTLS = false
	} else {
		usingTLS = true
	}

	rabbitCloseError = make(chan *amqp.Error)
	go rabbitConnector(clientMQTT)
	rabbitCloseError <- amqp.ErrClosed

	if token := clientMQTT.Subscribe(*subQueueTopic, 0, mqttCallback); token.Wait() && token.Error() != nil {
		log.Fatalf("Error subscribing to MQTT topic: %s", token.Error())
	}

	<-c

}

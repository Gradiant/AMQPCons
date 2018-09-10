# AMQP publisher

Module that consumes AMQP messages by connecting to queues

### Version

0.1

### Changelog

#### 0.1

- Initial version

### Dependencies

- MQTT client https://github.com/eclipse/paho.mqtt.golang
- AMQP client: https://github.com/streadway/amqp
- UUID generator: https://github.com/satori/go.uuid
- A MQTT broker like Mosquitto is needed to talk to other services

### Installation

- sudo docker build -t amqpcons .
- sudo docker run --name amqpcons -d amqpcons -uri user:pwd@amqp:5672/vhost -mqttBroker=mqtt:1883 -queueTopic=consumer/queue -connTopic=consumer/conn -dataTopic=consumer/data -cacert=./cacert.pem -clcert=./cert.pem -clkey=./key.pem

### Parameters

- uri: AMQP uri.
- mqttBroker: URL where the MQTT broker is located.
- queueTopic: MQTT topic where this service listens for messages to configure consumers
- connTopic: MQTT topic where this service publishes updates of the status of the connection.
- dataTopic: MQTT topic where this service publishes data received through the configured consumers.
- cacert: Path to the CA certificate.
- clcert: Path to the client certificate.
- clkey: Path to the client key.

### How it works

- If certificates and keys are provided the service tries to connect using TLS. Otherwise it connects without encryption.
- When providing certificates for TLS connection make sure that you update the needed files in the certs folder before starting the image.
- The service listens in queueTopic for messages to create or delete new consumers. The format of the message shall be in JSON as '{"Type":"NewConsumer","Queue":"queue_name","DeclareQueue":false}' if we want to create a consumer from a queue that is already configured in the broker and '{"Type":"NewConsumer","Exchange","exchange_name","ExchangeType":"topic","BindingKey":"key","Queue":"queue_name","DeclareQueue":true}' if we want to create in the broker all the needed actors to get messages. If the exchange already exists in the broker the consumer will create the queue and bind it to the existing exchange. When the consumer is created without errors the service sends a UUID used as consumer tag to identify the consumer.
- Each time a consumer receives a message the service will send its content through the dataTopic embedded in a JSON message as follows: {"ctag":"ctag_uuid","payload":"message payload"}
- To delete a consumer send a message with the following JSON format to the queueTopic: '{"Type":"DelConsumer","Ctag":"ctag_uuid"}'
- When the service detects a loss of connection it notifies it through connTopic by sending a '0'. It tries to reconnect to the other endpoint each 5 seconds.
- When the connection is recovered a '1' is sent through connTopic.

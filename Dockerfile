# Build stage
FROM resin/beaglebone-black-golang AS build-env
ADD . /src
RUN cd /src && go get -u github.com/eclipse/paho.mqtt.golang && go get -u github.com/streadway/amqp && CGO_ENABLED=0 GOOS=linux go build -o amqppub .

# Final stage
FROM hypriot/rpi-alpine-scratch
WORKDIR /app
RUN mkdir /app/certs
COPY certs/ /app/certs
COPY --from=build-env /src/amqppub /app/
ENTRYPOINT ["./amqppub"]

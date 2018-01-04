package main

import (
	"log"
	"github.com/streadway/amqp"
	"os"
	. "3pc/commons"
)

var state int // state of cohort
var conn *amqp.Connection
var ch *amqp.Channel
var coordQueue amqp.Queue

func initAmqp() {
	conn, err := amqp.Dial("amqp://localhost:5672/")
	FailOnError(err, "Failed to connect to RabbitMQ")
	ch, err = conn.Channel()
	FailOnError(err, "Failed to open a channel")
}

// queue for cohorts messages
func initCoordQueue() {
	q, err := ch.QueueDeclare(
		"coord", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	coordQueue = q
	FailOnError(err, "Failed to declare a queue")
}

// coord to cohorts publish-subscribe exchange
func initCoordExchange() {
	err := ch.ExchangeDeclare(
		"coordBC", // name
		"fanout",  // type
		true,      // durable
		false,     // auto-deleted
		false,     // internal
		false,     // no-wait
		nil,       // arguments
	)
	FailOnError(err, "Failed to declare an exchange")

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when usused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	FailOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name,    // queue name
		"",        // routing key
		"coordBC", // exchange
		false,
		nil)
	FailOnError(err, "Failed to bind a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	FailOnError(err, "Failed to register a consumer")

	go func() {
		for d := range msgs {
			receivedMsg(string(d.Body))
		}
	}()
}

func receivedMsg(msg string) {
	log.Printf("Received a message: %s", msg)

	switch {
	case state == Q && msg == COMMIT_REQ:
		state = W
		sendToCoord(AGREE)
	case state == W && msg == ABORT:
		state = A
	case state == W && msg == PREPARE:
		state = P
		sendToCoord(ACK)
	case state == P && msg == COMMIT:
		state = C
	}
}

func sendToCoord(body string) {
	err := ch.Publish(
		"",              // exchange
		coordQueue.Name, // routing key
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	log.Printf(" [x] Sent %s", body)
	FailOnError(err, "Failed to send a message")
}

// if first arg is 'T' then this coord send transaction request
func main() {

	initAmqp()
	defer conn.Close()
	defer ch.Close()
	initCoordQueue()
	initCoordExchange()

	if len(os.Args) > 1 {
		if os.Args[1] == "T" {
			sendToCoord(TR_REQ)
		}
	}

	select {}
}

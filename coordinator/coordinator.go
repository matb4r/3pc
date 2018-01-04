package main

import (
	"log"
	"github.com/streadway/amqp"
	"os"
	"strconv"
	. "3pc/commons"
)

var N int      // number of cohorts
var state int  // state of coordinator
var agreed = 0 //number of cohorts that agreed
var acked = 0  // number of cohorts that sent ack
var conn *amqp.Connection
var ch *amqp.Channel

func initAmqp() {
	conn, err := amqp.Dial("amqp://localhost:5672/")
	FailOnError(err, "Failed to connect to RabbitMQ")
	ch, err = conn.Channel()
	FailOnError(err, "Failed to open a channel")
}

// queue for cohorts messages
func initCoordQueue() {
	coordQueue, err := ch.QueueDeclare(
		"coord", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	FailOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		coordQueue.Name, // queue
		"",              // consumer
		true,            // auto-ack
		false,           // exclusive
		false,           // no-local
		false,           // no-wait
		nil,             // args
	)
	FailOnError(err, "Failed to register a consumer")

	// listen
	go func() {
		for d := range msgs {
			receivedMsg(string(d.Body))
		}
	}()
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
}

func publishToCohorts(body string) {
	err := ch.Publish(
		"coordBC", // exchange
		"",        // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	FailOnError(err, "Failed to publish a message")

	log.Printf(" [x] Sent %s", body)
}

func receivedMsg(msg string) {
	log.Printf("Received a message: %s", msg)

	switch {
	case state == Q && msg == TR_REQ:
		state = W
		publishToCohorts(COMMIT_REQ)
	case state == W && msg == AGREE:
		agreed += 1
		if agreed == N {
			agreed = 0
			state = P
			publishToCohorts(PREPARE)
		}
	case state == W && msg == ABORT:
		agreed = 0
		state = A
		publishToCohorts(ABORT)
	case state == P && msg == ACK:
		acked += 1
		if acked == N {
			acked = 0
			state = C
			publishToCohorts(COMMIT)
		}
	}
}

// first arg is expected number of cohorts
func main() {
	if len(os.Args) < 2 {
		log.Fatalln("First argument should be a expected number of cohorts")
	}
	n, err := strconv.Atoi(os.Args[1])
	FailOnError(err, "Wrong program arg")
	N = n

	initAmqp()
	defer conn.Close()
	defer ch.Close()
	initCoordQueue()
	initCoordExchange()

	select {}
}

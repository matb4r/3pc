package main

import (
	"log"
	"github.com/streadway/amqp"
	"os"
	. "3pc/commons"
	"bufio"
	"io/ioutil"
	"strings"
	"fmt"
)

var url string
var port string
var state int // state of cohort
var conn *amqp.Connection
var ch *amqp.Channel
var coordQueue amqp.Queue
var abortTransaction = false
var canCommit chan bool

func initAmqp() {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s/", url, port))
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
	log.Printf("+ Received: %s", msg)

	switch {
	case state == Q && msg == COMMIT_REQ:
		if abortTransaction {
			state = A
			sendToCoord(ABORT)
		} else {
			state = W
			sendToCoord(AGREE)
		}
	case state == W && msg == ABORT:
		state = A
		canCommit <- false
	case state == W && msg == PREPARE:
		state = P
		sendToCoord(ACK)
	case state == P && msg == ABORT:
		state = A
		canCommit <- false
	case state == P && msg == COMMIT:
		state = C
		canCommit <- true
	}

	log.Printf("= %d", state)
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
	log.Printf("- Sent: %s", body)
	FailOnError(err, "Failed to send a message")
}

func readStdin() string {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	FailOnError(err, "Failed to read an input")
	input = strings.Replace(input, "\n", "", -1) // convert CRLF to LF
	return input
}

func action(input string) {
	words := strings.Split(input, " ")
	switch words[0] {
	case "write":
		writeToFile(words[1], strings.Join(words[2:], " "))
	case "abort":
		abortTransaction = true
	}
}

func writeToFile(path string, text string) {
	canCommit = make(chan bool)

	sendToCoord(TR_REQ)

	if <-canCommit {
		err := ioutil.WriteFile(path, []byte(text), 0644)
		FailOnError(err, "Failed to write to a file")
		log.Printf("Write to %s success", path)
	} else {
		log.Printf("Write to %s aborted", path)
	}
}

// Args: url port
func parseProgramArgs() {
	if len(os.Args) < 3 {
		log.Fatalln("Usage: url port")
	}
	url = os.Args[1]
	port = os.Args[2]
}

func main() {
	parseProgramArgs()

	defer conn.Close()
	defer ch.Close()

	initAmqp()
	initCoordQueue()
	initCoordExchange()

	input := readStdin()
	action(input)

	select {}
}

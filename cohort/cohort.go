package main

import (
	. "github.com/matb4r/3pc/commons"
	"github.com/streadway/amqp"
	"log"
	"os"
	"bufio"
	"io/ioutil"
	"strings"
	"fmt"
	"time"
)

var url string
var port string
var state = Q // state of cohort
var conn *amqp.Connection
var ch *amqp.Channel
var coordQueue amqp.Queue
var abortTransaction = false
var canCommit chan bool
var timer *time.Timer
var timeout time.Duration = 3
var writing bool  // is this cohort wants to write to a file
var failure1 bool // simulate failure
var failure2 bool // simulate failure

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
	//bufio.NewReader(os.Stdin).ReadBytes('\n')
	switch {
	case state == Q && msg == COMMIT_REQ:
		if abortTransaction {
			state = A
			sendToCoord(ABORT)
		} else {
			state = W
			if failure1 {
				os.Exit(1)
			}
			sendToCoord(AGREE)
			timer = time.AfterFunc(time.Second*timeout, timeoutFunc)
		}
	case state == W && msg == ABORT:
		state = A
		registerCommit(false)
		exitProgram()
	case state == W && msg == PREPARE:
		state = P
		if failure2 {
			os.Exit(1)
		}
		sendToCoord(ACK)
		timer.Reset(time.Second * timeout)
	case state == P && msg == ABORT:
		state = A
		registerCommit(false)
		timer.Stop()
		exitProgram()
	case state == P && msg == COMMIT:
		state = C
		registerCommit(true)
		exitProgram()
	}

	log.Printf("[%s]", state)
}

func registerCommit(doCommit bool) {
	if writing {
		canCommit <- doCommit
	}
}

func exitProgram() {
	timer.Stop()
	if !writing {
		log.Printf("[%s]", state)
		os.Exit(0)
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
	log.Printf("- Sent: %s", body)
	FailOnError(err, "Failed to send a message")
}

func timeoutFunc() {
	log.Println("Timeout")
	if state == W {
		state = A
		canCommit <- false
	} else if state == P {
		state = C
		canCommit <- true
	}
}

func readStdin() string {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	FailOnError(err, "Failed to read an input")
	input = strings.Replace(input, "\n", "", -1) // convert CRLF to LF
	return input
}

func writeToFile(path string, text string) {
	canCommit = make(chan bool)
	writing = true

	sendToCoord(TR_REQ)

	if <-canCommit {
		err := ioutil.WriteFile(path, []byte(text), 0644)
		FailOnError(err, "Failed to write to a file")
		log.Printf("Write to \"%s\" to %s success", text, path)
	} else {
		log.Printf("Write to %s aborted", path)
	}
	os.Exit(0)
}

func parseConnectionArgs() {
	if len(os.Args) < 3 {
		log.Fatalln("Usage: url port ['write' path content / 'abort' / 'failure1' / 'failure2']")
	}
	url = os.Args[1]
	port = os.Args[2]
}

func parseActionArgs() {
	if len(os.Args) > 3 {
		switch os.Args[3] {
		case "write":
			writeToFile(os.Args[4], strings.Join(os.Args[5:], " "))
		case "abort":
			abortTransaction = true
		case "failure1":
			failure1 = true
		case "failure2":
			failure2 = true
		}
	}
}

func main() {
	parseConnectionArgs()

	defer conn.Close()
	defer ch.Close()
	initAmqp()
	initCoordQueue()
	initCoordExchange()

	parseActionArgs()

	select {}
}

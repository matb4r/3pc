package commons

import "log"

// states
const (
	Q int = iota
	W
	A
	P
	C
)

// types of messages
const (
	TR_REQ     string = "transaction_request"
	COMMIT_REQ string = "commit_request"
	ABORT      string = "abort"
	AGREE      string = "agree"
	PREPARE    string = "prepare"
	COMMIT     string = "commit"
	ACK        string = "ack"
)

func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

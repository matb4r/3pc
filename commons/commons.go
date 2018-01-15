package commons

import "log"

// states
const (
	Q string = "Q"
	W        = "W"
	A        = "A"
	P        = "P"
	C        = "C"
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

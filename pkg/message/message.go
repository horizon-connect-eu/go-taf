// Produces messages that contain an ID (an integer from 0 to 10), a value (an integer from 100 to 999), and a string (either "TSM" or "TMM").
// The messages are in form of a struct of type message.

package message

import (
	"math/rand/v2"
)

type Message struct {
	ID    int
	Value int
	Rx    string
}

func New(Tx_ID int, Tx_value int, Tx_Rx string) Message {

	msg1 := Message{
		Tx_ID,
		Tx_value,
		Tx_Rx,
	}

	return msg1
}

func Generate() Message {
	var Tx_Rx string
	Tx_ID := rand.IntN(10)
	Tx_value := 100 + rand.IntN(899)
	if rand.IntN(99)%2 == 0 {
		Tx_Rx = "TMM"
	} else {
		Tx_Rx = "TSM"
	}

	msg := New(Tx_ID, Tx_value, Tx_Rx)

	return msg
}

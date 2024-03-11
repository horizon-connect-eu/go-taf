// Produces messages that contain an ID (an integer from 0 to 10), a value (an integer from 100 to 999), and a string (either "TSM" or "TMM").
// The messages are in form of a struct of type message.

package msgGenerator

import (
	"fmt"
	"math/rand/v2"
)

type message struct {
	ID    int
	value int
	Rx    string
}

func msg_Tx(Tx_ID int, Tx_value int, Tx_Rx string) message {

	msg1 := message{
		Tx_ID,
		Tx_value,
		Tx_Rx,
	}

	return msg1
}

func msgGenerator() {

	var Tx_Rx string
	Tx_ID := rand.IntN(10)
	Tx_value := 100 + rand.IntN(899)
	if rand.IntN(99)%2 == 0 {
		Tx_Rx = "TMM"
	} else {
		Tx_Rx = "TSM"
	}

	msg := msg_Tx(Tx_ID, Tx_value, Tx_Rx)
	fmt.Println("Message: ", msg)
}

package message

import (
	"math/rand/v2"
)

type Message struct {
	ID    int
	Value int
	Rx    string
	Type  string
}

// Creates a new message with the given parameters
func New(Tx_ID int, Tx_value int, Tx_Rx string, Type string) Message {
	msg1 := Message{
		Tx_ID,
		Tx_value,
		Tx_Rx,
		Type,
	}

	return msg1
}

// Generates a new message with random values
//
// Produces messages that contain an ID (an integer from 0 to 10), a value (an integer from 100 to 999), and a string (either "TSM" or "TMM").
func Generate() Message {
	var Tx_Rx string
	Tx_ID := rand.IntN(100000)
	Tx_value := 100 + rand.IntN(899)
	if rand.IntN(99)%2 == 0 {
		Tx_Rx = "TMM"
	} else {
		Tx_Rx = "TSM"
	}

	var typeType string
	if Tx_ID%2 == 0 {
		typeType = "A"
	} else {
		typeType = "B"
	}

	msg := New(Tx_ID, Tx_value, Tx_Rx, typeType)

	return msg
}

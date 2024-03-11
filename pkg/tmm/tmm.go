package tmm

import (
	"fmt"

	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/message"
)

func Run(input chan message.Message, output chan message.Message) {
	for {
		received := <-input
		if received.Rx == "TMM" {
			fmt.Printf("I am TMM, received %+v\n", received)
			output <- received
		}
	}
}

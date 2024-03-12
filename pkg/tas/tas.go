package tas

import (
	"context"
	"fmt"
	"os"

	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/message"
)

func Run(ctx context.Context, input chan message.TasResponse, output chan message.TasQuery) {

	defer func() {
		//log.Println("TAS: shutting down")
	}()

	for counter := 0; ; counter++ {

		// Get user input
		fmt.Println("Enter the ID of the corresponding sum you want to request: [0-9]")
		var userInput int
		_, error := fmt.Fscanln(os.Stdin, &userInput)

		if error != nil {
			//log.Println("Error reading user input")
			return
		}

		if userInput < 0 || userInput > 9 {
			//log.Println("Invalid user input")
			return
		}

		output <- message.TasQuery{QueryID: counter, RequestedID: userInput}

		select {
		case <-ctx.Done():
			return
		case response := <-input:
			response = response //TODO
			//log.Printf("I am TAS, Received response: %+v\n", response)
		}
	}
}

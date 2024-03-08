package main

import (
	"fmt"

	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/internal/somepackage"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/pubapi"
)

func main() {
	fmt.Println("Hello from main.go")
	somepackage.SomeApiFunction()

	result := pubapi.Add(1, 3)
	fmt.Println("Result of calculation: ", result)
}

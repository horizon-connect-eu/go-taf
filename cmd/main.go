// # main package
//
// TODO
package main

import (
	"context"
	"fmt"
	"github.com/pterm/pterm"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vs-uulm/go-taf/pkg/evidencecollection"

	"github.com/vs-uulm/go-taf/pkg/config"
	"github.com/vs-uulm/go-taf/pkg/message"
	"github.com/vs-uulm/go-taf/pkg/trustassessment"
	"github.com/vs-uulm/go-taf/pkg/trustmodel"
	"github.com/vs-uulm/go-taf/pkg/trustsource"
)

//go:generate go run ../plugins/plugins.go

// Blocks until the process receives SIGTERM (or equivalent).
func WaitForCtrlC() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}

// main is the entry point of the application.
// It starts all the components of the application and waits for a signal to stop the application.
func main() {
	tafConfig := config.DefaultConfig
	// First, see whether a config file path has been specified
	if filepath, ok := os.LookupEnv("TAF_CONFIG"); ok {
		var err error
		tafConfig, err = config.LoadJSON(filepath)
		if err != nil {
			log.Fatalf("main: error reading config file %s: %s\n", filepath, err.Error())
		}
	}
	log.Printf("Running with configuration: %+v\n", tafConfig)

	//Create main channels
	//c1 := make(chan message.InternalMessage, tafConfig.ChanBufSize)
	c2 := make(chan message.InternalMessage, tafConfig.ChanBufSize)

	//c3 := make(chan message.InternalMessage, tafConfig.ChanBufSize)
	//c4 := make(chan message.InternalMessage, tafConfig.ChanBufSize)

	tmm2tamChannel := make(chan trustassessment.Command, tafConfig.ChanBufSize)
	eci2tsm := make(chan message.EvidenceCollectionMessage, tafConfig.ChanBufSize)
	tsm2tamChannel := make(chan trustassessment.Command, tafConfig.ChanBufSize)

	tmts := map[string]int{}

	ctx := context.Background()
	ctx, cancelFunc := context.WithCancel(ctx)
	defer time.Sleep(1 * time.Second) // TODO replace this cleanup interval with waitgroups
	defer cancelFunc()

	//	go v2xlistener.Run(ctx, tafConfig.V2X, []chan message.InternalMessage{c1, c2})

	evidenceCollection, err := evidencecollection.New(eci2tsm, tafConfig)
	if err != nil {
		log.Fatal(err)
	}
	go evidenceCollection.Run(ctx)

	trustAssessmentManager, err := trustassessment.NewManager(tafConfig, tmts)
	if err != nil {
		log.Fatal(err)
	}
	go trustAssessmentManager.Run(ctx, tmm2tamChannel, tsm2tamChannel)

	go trustmodel.Run(ctx, tmm2tamChannel)
	go trustsource.Run(ctx, c2, eci2tsm, tsm2tamChannel)

	/*
		ticker := time.NewTicker(1 * time.Second)
		for range ticker.C {
			fmt.Println("CHANNELS: ", len(c1), len(c2), len(c3), len(c4))
		}
	*/

	go printOutput()

	WaitForCtrlC()

}

func renderTable(data pterm.TableData) {
	pterm.DefaultTable.WithHasHeader().WithBoxed().WithRightAlignment().WithData(data).Render()
}

func clearTerminal() {
	fmt.Print("\033[H\033[2J")
}

func printOutput() {
	//	headerStyle := pterm.NewStyle(pterm.Bold, pterm.BgBlack, pterm.FgWhite)
	tableData := pterm.TableData{
		{"Rel. ID", "Trustor", "Trustee", "ω", "Trust Decision"},
		{"4711-123", "TAF", "ECU1", "(0.1, 0.2, 0.3, 0.4)", pterm.Green(" ✔ ")},
		{"4711-124", "TAF", "ECU2", "(0.1, 0.2, 0.3, 0.4)", pterm.Green(" ✔ ")},
	}

	// Render the initial table
	//clearTerminal()
	renderTable(tableData)

	// Simulate a delay before updating the table
	time.Sleep(2 * time.Second)

	// Update the table data
	tableData[1][4] = pterm.Red(" ✗ ")

	// Clear the terminal and re-render the table with updated data
	//clearTerminal()
	renderTable(tableData)
}

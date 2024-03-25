package v2xlistener

import (
	"context"
	"fmt"
	"time"

	"gitlab-vs.informatik.uni-ulm.de/connect/taf-brussels-demo/pkg/config"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-brussels-demo/pkg/message"
)

func Run(ctx context.Context, v2xconfig config.V2XConfiguration, outputs []chan message.InternalMessage) {
	defer func() {
		//log.Println("V2XListener: shutting down")
	}()

	msgCtr := 0
	lastTime := time.Now()

	// Ticker for measuring throughput
	bmTicker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return
		case <-bmTicker.C:
			delta := time.Since(lastTime)
			genRate := float64(msgCtr) / delta.Seconds()
			fmt.Printf("v2x: %e messages per second\n", genRate)
			msgCtr = 0
			lastTime = time.Now()
		default:
			msg := message.Generate()
			sendToAll(outputs, msg)
			msgCtr++
			time.Sleep(time.Duration(v2xconfig.SendIntervalNs) * time.Nanosecond)
		}
	}
}

// Write a message to all channels in a slice of channels.
func sendToAll(outputs []chan message.InternalMessage, msg message.InternalMessage) {
	// TODO: think about what should happen if a channel is full.
	// What if one is full and the other is not?
	// We should document this.
	for _, channel := range outputs {
		select {
		case channel <- msg:
		default:
			//fmt.Println("Channel full!")
			channel <- msg
		}
	}
}

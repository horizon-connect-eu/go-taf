package v2xlistener

import (
	"context"

	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/config"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/message"
)

func Run(ctx context.Context, v2xconfig config.V2XConfiguration, outputs []chan message.Message) {

	defer func() {
		//log.Println("V2XListener: shutting down")
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg := message.Generate()
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
			//time.Sleep(time.Duration(v2xconfig.SendIntervalMs))
		}
	}
}

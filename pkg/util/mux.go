package util

// Multiplex sourceA and sourceB into a single channel.
func Mux[T any](sourceA, sourceB <-chan T, sink chan<- T) {
	go func() {
		defer close(sink)
		for {
			select {
			case msg, ok := <-sourceA:
				if ok {
					sink <- msg
				} else {
					sourceA = nil
				}
			case msg, ok := <-sourceB:
				if ok {
					sink <- msg
				} else {
					sourceB = nil
				}
			}
			if sourceA == nil && sourceB == nil {
				break
			}
		}
	}()
}

// Multiplex multiple source channels into one output channel.
// ATTENTION: NOT FAIR! In case more than one channel is ready to be read
// from, the channels at higher indices of `sources` have an advantage of being selected.
// THEREFORE, DO NOT USE THIS FUNCTION IF FAIRNESS IS NEEDED.
// `sources` will never be written to.
func MuxMany[T any](sources []chan T, sink chan<- T) {
	var csink chan T = nil

	// TODO make this fair.
	// Possible references how this could be done:
	// - https://github.com/madlambda/spells/blob/fadce9f961a109e2723dd7a686151ff73f52057e/muxer/muxer.go#L30
	// - https://katcipis.github.io/blog/mux-channels-go/

	for i, source := range sources {
		if i == len(sources)-1 {
			Mux(csink, source, sink)
			//fmt.Println("A")
		} else {
			newSink := make(chan T)
			Mux(csink, source, newSink)
			csink = newSink
			//fmt.Println("B")
		}
	}

}

package util

import "sync"

// Multiplex multiple source channels (`sources`) into one output channel (`sink`).
// This function returns immediately.
// `sink` is closed after all channels in `sources` have been closed.
func Mux[T any](sink chan<- T, sources ...chan T) {
	// Copied from https://go.dev/blog/pipelines

	var wg sync.WaitGroup

	copier := func(c chan T) {
		for x := range c {
			sink <- x
		}
		wg.Done()
	}
	wg.Add(len(sources))
	for _, c := range sources {
		go copier(c)
	}
	go func() {
		wg.Wait()
		close(sink)
	}()
}

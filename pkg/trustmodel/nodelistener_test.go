package trustmodel

import (
	"sync"
	"testing"
	"time"
)

type Observer struct {
	id string
	T  *testing.T
}

func (o Observer) handleNodeAdded(id string) {
	o.T.Log("Observer " + o.id + " handles node added " + id + "\n")
}
func (o Observer) handleNodeRemoved(id string) {
	o.T.Log("Observer " + o.id + " handles node removed " + id + "\n")
}

func TestListener(t *testing.T) {

	listener := CreateListener(2, 1)

	obs1 := Observer{id: "1", T: t}
	obs2 := Observer{id: "2", T: t}

	listener.registerObserver(obs1)
	listener.registerObserver(obs2)

	var wg sync.WaitGroup
	wg.Add(5)

	go func() {
		listener.AddNode("test1")
		wg.Done()
	}()
	go func() {
		listener.AddNode("test2")
		wg.Done()
	}()
	go func() {
		time.Sleep(500 * time.Millisecond)
		listener.removeObserver(obs2)
		wg.Done()
	}()
	go func() {
		time.Sleep(1000 * time.Millisecond)
		listener.RemoveNode("test2")
		wg.Done()
	}()
	go func() {
		time.Sleep(4000 * time.Millisecond)
		wg.Done()
	}()

	wg.Wait()
}

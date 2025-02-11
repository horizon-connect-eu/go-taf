package trustmodel

import (
	"sync"
	"time"
)

/*
EntityObserver implements the observer pattern and provides an interface to register listeners to be called when new
entities have been added or removed.
*/
type EntityObserver struct {
	nodes     map[string]int64
	observers map[observer]bool
	lock      *sync.RWMutex
	ttl       int
}

type observer interface {
	handleNodeAdded(identifier string)
	handleNodeRemoved(identifier string)
}

func CreateListener(ttlSeconds int, checkIntervalSeconds int) EntityObserver {
	listener := EntityObserver{
		nodes:     make(map[string]int64),
		observers: make(map[observer]bool),
		lock:      &sync.RWMutex{},
	}
	go func() {
		for now := range time.Tick(time.Duration(checkIntervalSeconds) * time.Second) {
			listener.lock.Lock()
			for key, timestamp := range listener.nodes {
				if (now.Unix()) > timestamp+int64(ttlSeconds) {
					delete(listener.nodes, key)
					listener.notifyObserversOnNodeRemoved(key)
				}
			}
			listener.lock.Unlock()
		}
	}()
	return listener
}

func (l *EntityObserver) registerObserver(observer observer) {
	l.observers[observer] = true
}

func (l *EntityObserver) removeObserver(observer observer) {
	delete(l.observers, observer)
}

func (l *EntityObserver) notifyObserversOnNodeAdded(identifier string) {
	for observer, _ := range l.observers {
		observer.handleNodeAdded(identifier)
	}
}

func (l *EntityObserver) notifyObserversOnNodeRemoved(identifier string) {
	for observer, _ := range l.observers {
		observer.handleNodeRemoved(identifier)
	}
}

func (l *EntityObserver) AddNode(identifier string) {
	l.lock.Lock()
	defer l.lock.Unlock()

	_, exists := l.nodes[identifier]
	l.nodes[identifier] = time.Now().Unix()
	if !exists {
		l.notifyObserversOnNodeAdded(identifier)
	}
}

func (l *EntityObserver) RemoveNode(identifier string) {
	l.lock.Lock()
	defer l.lock.Unlock()

	_, exists := l.nodes[identifier]
	if exists {
		delete(l.nodes, identifier)
		l.notifyObserversOnNodeRemoved(identifier)
	}
}

func (l *EntityObserver) Nodes() []string {
	l.lock.RLock()
	defer l.lock.RUnlock()

	nodes := make([]string, len(l.nodes))
	i := 0
	for node := range l.nodes {
		nodes[i] = node
		i++
	}
	return nodes
}

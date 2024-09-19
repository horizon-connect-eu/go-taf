package trustmodel

import (
	"sync"
	"time"
)

type v2xObserver struct {
	nodes     map[string]int64
	observers map[observer]bool
	lock      *sync.RWMutex
	ttl       int
}

type subject interface {
	registerObserver(observer observer)
	removeObserver(observer observer)
	notifyObserversOnNodeAdded(identifier string)
	notifyObserversOnNodeRemoved(identifier string)
}

type observer interface {
	handleNodeAdded(identifier string)
	handleNodeRemoved(identifier string)
}

func CreateListener(ttlSeconds int, checkIntervalSeconds int) v2xObserver {
	listener := v2xObserver{
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

func (l *v2xObserver) registerObserver(observer observer) {
	l.observers[observer] = true
}

func (l *v2xObserver) removeObserver(observer observer) {
	delete(l.observers, observer)
}

func (l v2xObserver) notifyObserversOnNodeAdded(identifier string) {
	for observer, _ := range l.observers {
		observer.handleNodeAdded(identifier)
	}
}

func (l v2xObserver) notifyObserversOnNodeRemoved(identifier string) {
	for observer, _ := range l.observers {
		observer.handleNodeRemoved(identifier)
	}
}

func (l *v2xObserver) AddNode(identifier string) {
	l.lock.Lock()
	defer l.lock.Unlock()

	_, exists := l.nodes[identifier]
	l.nodes[identifier] = time.Now().Unix()
	if !exists {
		l.notifyObserversOnNodeAdded(identifier)
	}
}

func (l *v2xObserver) RemoveNode(identifier string) {
	l.lock.Lock()
	defer l.lock.Unlock()

	_, exists := l.nodes[identifier]
	if exists {
		delete(l.nodes, identifier)
		l.notifyObserversOnNodeRemoved(identifier)
	}
}

func (l *v2xObserver) Nodes() []string {
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

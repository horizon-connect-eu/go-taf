package web

import (
	"github.com/vs-uulm/go-taf/pkg/listener"
)

func (s *Webserver) OnSessionCreated(event listener.SessionCreatedEvent) {
	go func() {
		s.listenerChannel <- event
	}()
}

func (s *Webserver) OnSessionTorndown(event listener.SessionDeletedEvent) {
	go func() {
		s.listenerChannel <- event
	}()
}

func (s *Webserver) OnATLUpdated(event listener.ATLUpdatedEvent) {
	go func() {
		s.listenerChannel <- event
	}()

}

func (s *Webserver) OnATLRemoved(event listener.ATLRemovedEvent) {
	go func() {
		s.listenerChannel <- event
	}()
}

func (s *Webserver) OnTrustModelInstanceSpawned(event listener.TrustModelInstanceSpawnedEvent) {
	go func() {
		s.listenerChannel <- event
	}()
}

func (s *Webserver) OnTrustModelInstanceUpdated(event listener.TrustModelInstanceUpdatedEvent) {
	go func() {
		s.listenerChannel <- event
	}()
}

func (s *Webserver) OnTrustModelInstanceDeleted(event listener.TrustModelInstanceDeletedEvent) {
	go func() {
		s.listenerChannel <- event
	}()

}

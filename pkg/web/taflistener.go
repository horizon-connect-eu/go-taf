package web

import (
	"github.com/vs-uulm/go-taf/pkg/listener"
)

func (s *Webserver) OnSessionCreated(event listener.SessionCreatedEvent) {
	s.listenerChannel <- event
}

func (s *Webserver) OnSessionTorndown(event listener.SessionDeletedEvent) {
	s.listenerChannel <- event
}

func (s *Webserver) OnATLUpdated(event listener.ATLUpdatedEvent) {
	s.listenerChannel <- event
}

func (s *Webserver) OnATLRemoved(event listener.ATLRemovedEvent) {
	s.listenerChannel <- event
}

func (s *Webserver) OnTrustModelInstanceSpawned(event listener.TrustModelInstanceSpawnedEvent) {
	s.listenerChannel <- event
}

func (s *Webserver) OnTrustModelInstanceUpdated(event listener.TrustModelInstanceUpdatedEvent) {
	s.listenerChannel <- event
}

func (s *Webserver) OnTrustModelInstanceDeleted(event listener.TrustModelInstanceDeletedEvent) {
	s.listenerChannel <- event
}

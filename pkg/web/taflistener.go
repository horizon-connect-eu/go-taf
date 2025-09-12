package web

import (
	"github.com/horizon-connect-eu/go-taf/pkg/listener"
)

/*
This file contains all listener functions that will be called by the core TAF whenever a certain event occurs.
To decouple processing from the core TAF, the event is passed to the listener channel and then handled from the
web server go-routine, separated from the core TAF.
*/

func (s *Webserver) OnSessionCreated(event listener.SessionCreatedEvent) {
	s.listenerChannel <- event
}

func (s *Webserver) OnSessionTorndown(event listener.SessionTorndownEvent) {
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

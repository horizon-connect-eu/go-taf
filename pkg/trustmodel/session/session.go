package session

import "github.com/vs-uulm/go-taf/pkg/trustmodel/trustmodelinstance"

type Session interface {
	ID() string
	TrustModelInstances() map[string]trustmodelinstance.TrustModelInstance
	Client() string
}

type Instance struct {
	id     string
	tMIs   map[string]trustmodelinstance.TrustModelInstance
	client string
}

func NewInstance(id, client string) Session {
	return &Instance{
		id:     id,
		tMIs:   make(map[string]trustmodelinstance.TrustModelInstance),
		client: client,
	}
}

func (d *Instance) ID() string {
	return d.id
}
func (d *Instance) TrustModelInstances() map[string]trustmodelinstance.TrustModelInstance {
	return d.tMIs
}
func (d *Instance) Client() string {
	return d.client
}

package session

import (
	"github.com/vs-uulm/go-taf/pkg/core"
)

type Session interface {
	ID() string
	TrustModelInstances() map[string]core.TrustModelInstance
	Client() string
}

type Instance struct {
	id     string
	tMIs   map[string]core.TrustModelInstance
	client string
}

func NewInstance(id, client string) Session {
	return &Instance{
		id:     id,
		tMIs:   make(map[string]core.TrustModelInstance),
		client: client,
	}
}

func (d *Instance) ID() string {
	return d.id
}
func (d *Instance) TrustModelInstances() map[string]core.TrustModelInstance {
	return d.tMIs
}
func (d *Instance) Client() string {
	return d.client
}

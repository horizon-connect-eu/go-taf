package session

import "github.com/vs-uulm/go-taf/pkg/core"

type State uint8

const (
	NON_EXISTENT State = iota
	INITIALIZING
	ESTABLISHED
	TEARING_DOWN
	TORN_DOWN
)

type Session interface {
	ID() string
	TrustModelInstances() map[string]bool
	TrustModelTemplate() core.TrustModelTemplate
	Client() string
	HasTMI(tmiID string) bool
	State() State
	Established()
	TearingDown()
	TornDown()
	AddSubscription(identifier string)
	ListSubscriptions() []string
	RemoveSubscription(identifier string)
}

type Instance struct {
	id            string
	tMIs          map[string]bool
	client        string
	state         State
	tmt           core.TrustModelTemplate
	subscriptions map[string]bool
}

func NewInstance(id, client string, tmt core.TrustModelTemplate) Session {
	return &Instance{
		id:            id,
		tMIs:          make(map[string]bool),
		subscriptions: make(map[string]bool),
		client:        client,
		tmt:           tmt,
		state:         INITIALIZING,
	}
}

func (s *Instance) ID() string {
	return s.id
}

func (s *Instance) TrustModelInstances() map[string]bool {
	return s.tMIs
}

func (s *Instance) TrustModelTemplate() core.TrustModelTemplate {
	return s.tmt
}

func (s *Instance) Client() string {
	return s.client
}

func (s *Instance) State() State {
	return s.state
}

func (s *Instance) Established() {
	s.state = ESTABLISHED
}

func (s *Instance) TearingDown() {
	s.state = TEARING_DOWN
}

func (s *Instance) TornDown() {
	s.state = TORN_DOWN
}

func (s *Instance) HasTMI(tmiID string) bool {
	val, exists := s.tMIs[tmiID]
	if !exists {
		return false
	}
	return val
}

func (s *Instance) AddSubscription(identifier string) {
	s.subscriptions[identifier] = true
}

func (s *Instance) RemoveSubscription(identifier string) {
	_, exists := s.subscriptions[identifier]
	if exists {
		delete(s.subscriptions, identifier)
	}
}

func (s *Instance) ListSubscriptions() []string {
	subs := make([]string, len(s.subscriptions))
	i := 0
	for k := range s.subscriptions {
		subs[i] = k
		i++
	}
	return subs
}

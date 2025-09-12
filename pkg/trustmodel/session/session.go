package session

import "github.com/horizon-connect-eu/go-taf/pkg/core"

/*
State specifies the state a Session is in.
*/
type State uint8

const (
	NON_EXISTENT State = iota
	INITIALIZING
	ESTABLISHED
	TEARING_DOWN
	TORN_DOWN
)

/*
A Session is stateful connection between a client and the TAF based upon a specified type of trust model.
TrustModelInstances can only be created as part of a Session, and the existence of session is necessary to start
the collection of evidence.
*/
type Session interface {

	/*
		ID returns the TAF-internal identifier of this Session.
	*/
	ID() string

	/*
		TrustModelInstances returns a map of TMIs associated with this session. In this map, the key are the
		short TMI IDs, the value are the corresponding full TMI IDs.
	*/
	TrustModelInstances() map[string]string

	/*
		TrustModelTemplate returns the TMT used for this Session.
	*/
	TrustModelTemplate() core.TrustModelTemplate

	/*
		Client returns the identifier of the client the Session belongs to.
	*/
	Client() string

	/*
		HasTMI checks whether the provided (short) TMI-ID is present in this Session.
	*/
	HasTMI(tmiID string) bool

	/*
		State returns the State this Session is in.
	*/
	State() State

	/*
		Established transitions this Session into the ESTABLISHED state.
	*/
	Established()

	/*
		TearingDown transitions this Session into the TEARING_DOWN state.

	*/
	TearingDown()

	/*
		TornDown transitions this Session into the TORN_DOWN state.

	*/
	TornDown()

	/*
		Adds a TAS subscription to the Session based on the subscription identifier.
	*/
	AddSubscription(identifier string)

	/*
		Lists the subscription idenitifers for all subscriptions of this Session.
	*/
	ListSubscriptions() []string

	/*
		Removes a TAS subscription from the Session based on the subscription identifier.
	*/
	RemoveSubscription(identifier string)

	/*
		SetDynamicSpawner sets the dynamic spawner for this Session, if available.
		Must only be called once at the beginning of session when calling Spawn() at the TrustModelTemplate.
	*/
	SetDynamicSpawner(spawner core.DynamicTrustModelInstanceSpawner)

	/*
		DynamicSpawner returns the dynamic spawner set for this Session, if available.
	*/
	DynamicSpawner() core.DynamicTrustModelInstanceSpawner

	/*
		SetTrustSourceQuantifiers sets the trust source quantifiers for this Session.
		Must only be called once at the beginning of session when calling Spawn() at the TrustModelTemplate.
	*/
	SetTrustSourceQuantifiers([]core.TrustSourceQuantifier)

	/*
		TrustSourceQuantifiers returns the list of core.TrustSourceQuantifier(s) set of this Session.
	*/
	TrustSourceQuantifiers() []core.TrustSourceQuantifier
}

type Instance struct {
	id            string
	tMIs          map[string]string
	client        string
	state         State
	tmt           core.TrustModelTemplate
	subscriptions map[string]bool
	spawner       core.DynamicTrustModelInstanceSpawner
	tsqs          []core.TrustSourceQuantifier
}

func NewInstance(id, client string, tmt core.TrustModelTemplate) Session {
	return &Instance{
		id:            id,
		tMIs:          make(map[string]string),
		subscriptions: make(map[string]bool),
		client:        client,
		tmt:           tmt,
		state:         INITIALIZING,
		spawner:       nil,
	}
}

func (s *Instance) ID() string {
	return s.id
}

func (s *Instance) TrustModelInstances() map[string]string {
	return s.tMIs
}

func (s *Instance) TrustModelTemplate() core.TrustModelTemplate {
	return s.tmt
}

func (s *Instance) Client() string {
	return s.client
}

func (s *Instance) SetDynamicSpawner(spawner core.DynamicTrustModelInstanceSpawner) {
	s.spawner = spawner
}
func (s *Instance) DynamicSpawner() core.DynamicTrustModelInstanceSpawner {
	return s.spawner
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
	_, exists := s.tMIs[tmiID]
	return exists
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

func (s *Instance) SetTrustSourceQuantifiers(tsqs []core.TrustSourceQuantifier) {
	s.tsqs = tsqs
}

func (s *Instance) TrustSourceQuantifiers() []core.TrustSourceQuantifier {
	return s.tsqs
}

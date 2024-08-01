package session

type Session interface {
	ID() string
	TrustModelInstances() map[string]bool
	Client() string
	HasTMI(tmiID string) bool
}

type Instance struct {
	id     string
	tMIs   map[string]bool
	client string
}

func NewInstance(id, client string) Session {
	return &Instance{
		id:     id,
		tMIs:   make(map[string]bool),
		client: client,
	}
}

func (s *Instance) ID() string {
	return s.id
}
func (s *Instance) TrustModelInstances() map[string]bool {
	return s.tMIs
}
func (s *Instance) Client() string {
	return s.client
}

func (s *Instance) HasTMI(tmiID string) bool {
	val, exists := s.tMIs[tmiID]
	if !exists {
		return false
	}
	return val
}

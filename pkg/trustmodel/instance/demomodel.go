package instance

type TrustModelInstance struct {
	id          int
	tmt         string
	omega1      float64
	omega2      float64
	version     int
	fingerprint int
}

func NewTrustModelInstance(id int, tmt string) TrustModelInstance {
	return TrustModelInstance{
		id:     id,
		tmt:    tmt,
		omega1: 0,
		omega2: 0,
	}
}

// TODO: Implement return hardcoded structure of this trust model instance
func (i *TrustModelInstance) getStructure() {

}

// TODO: Implement return values of this trust model instance
func (i *TrustModelInstance) getValues() {

}

func (i *TrustModelInstance) GetId() int {
	return i.id
}

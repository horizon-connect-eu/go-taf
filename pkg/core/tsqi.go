package core

type TrustSourceQuantifierInstance struct {
	Trustee  string
	Trustor  string
	Scope    string
	Evidence []Evidence
}

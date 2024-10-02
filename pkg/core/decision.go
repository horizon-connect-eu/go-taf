package core

/*
A TrustDecision represents the output of a trust decision engine execution.
*/
type TrustDecision uint8

const (
	/*
		Value representing "FALSE" as decision output.
	*/
	NOT_TRUSTWORTHY TrustDecision = iota

	/*
		Value representing "TRUE" as decision output.
	*/
	TRUSTWORTHY

	/*
		Value representing no decision, e.g., due to high uncertainty.
	*/
	UNDECIDABLE
)

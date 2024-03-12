package message

type TasQuery struct {
	QueryID     int
	RequestedID int
}

type TasResponse struct {
	ResponseID    int
	ResponseValue int
}

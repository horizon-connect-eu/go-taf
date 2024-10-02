package trustmodelupdate

import "github.com/vs-uulm/go-taf/pkg/core"

/*
RefreshCPM is an TMI update operation that updates the structure of a trust model according to the observations of a CPM message.
*/
type RefreshCPM struct {
	sourceID string
	objects  []string
}

func (r RefreshCPM) SourceID() string {
	return r.sourceID
}
func (r RefreshCPM) Objects() []string {
	return r.objects
}

func CreateRefreshCPM(sourceID string, objects []string) RefreshCPM {
	return RefreshCPM{
		sourceID: sourceID,
		objects:  objects,
	}
}

func (u RefreshCPM) Type() core.UpdateOp {
	return core.REFRESH_CPM
}

package trustassessment

import (
	"github.com/horizon-connect-eu/go-taf/pkg/core"
	"testing"
)

func TestTable(t *testing.T) {

	table := CreateTrustModelInstanceTable()

	client := "A"
	session := "B"
	tmt := "CACC@1.2.3"
	tmi := "4711"

	id := core.MergeFullTMIIdentifier(client, session, tmt, tmi)

	table.RegisterTMI(client, session, tmt, tmi)

	table.RegisterTMI("A", "X", "CACC@1.2.3", "123")
	table.RegisterTMI("A", "Y", "IMA@1.2.3", "798")
	table.RegisterTMI("B", "Z", "CACC@1.2.3", "456")

	t.Log(table.ExistsTMI(client, session, tmt, tmi))
	t.Log(table.ExistsTMI(client, session, tmt, "nope"))

	c, s, tm, i := core.SplitFullTMIIdentifier(id)

	t.Log("Client =", c)
	t.Log("Session =", s)
	t.Log("tmt =", tm)
	t.Log("tmi =", i)

	hits, _ := table.QueryTMIs("//A/*/*/*")
	t.Log(len(hits))
	for _, hit := range hits {
		t.Log(hit)
	}
	hits, _ = table.QueryTMIs("//*/*/CACC@1.2.3/*")
	t.Log(len(hits))
	for _, hit := range hits {
		t.Log(hit)
	}
}

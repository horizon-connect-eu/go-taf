package intersectionmovementassist

import (
	"github.com/vs-uulm/go-taf/pkg/core"

	"testing"
)

func TestLoadJson(t *testing.T) {
	param := map[core.EvidenceType]int{
		core.TCH_SECURE_BOOT:                          1,
		core.TCH_ACCESS_CONTROL:                       0,
		core.TCH_CONTROL_FLOW_INTEGRITY:               -1,
		core.TCH_SECURE_OTA:                           0,
		core.TCH_APPLICATION_ISOLATION:                1,
		core.TCH_CONFIGURATION_INTEGRITY_VERIFICATION: -1,
	}

	sl := trustSourceQuantifiers[0].Quantifier(param)

	t.Logf(sl.String())

	paramMBD := map[core.EvidenceType]int{
		core.MBD_MISBEHAVIOR_REPORT: 3,
	}

	sl2 := trustSourceQuantifiers[1].Quantifier(paramMBD)

	t.Logf(sl2.String())

}

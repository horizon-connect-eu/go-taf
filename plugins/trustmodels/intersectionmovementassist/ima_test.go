package intersectionmovementassist

import (
	"github.com/vs-uulm/go-taf/pkg/core"

	"testing"
)

func TestTrustSourceQuantifierFunctions(t *testing.T) {
	param := map[core.EvidenceType]int{
		core.TCH_SECURE_BOOT:                          1,
		core.TCH_ACCESS_CONTROL:                       0,
		core.TCH_CONTROL_FLOW_INTEGRITY:               -1,
		core.TCH_SECURE_OTA:                           0,
		core.TCH_APPLICATION_ISOLATION:                1,
		core.TCH_CONFIGURATION_INTEGRITY_VERIFICATION: -1,
	}

	trustSourceQuantifiers, _ := createTrustSourceQuantifiers(nil)
	sl := trustSourceQuantifiers[0].Quantifier(param)

	t.Logf(sl.String())

	paramMBD := map[core.EvidenceType]int{
		core.MBD_MISBEHAVIOR_REPORT: 3,
	}

	sl2 := trustSourceQuantifiers[1].Quantifier(paramMBD)

	t.Logf(sl2.String())

	params := make(map[string]string)
	params["MBD_ND_SPEE_CONS"] = "2"
	params["MBD_D_SPEE_PLAU"] = "2"
	params["TCH_EXISTENCE_SECURE_BOOT"] = "3"
	params["TCH_OUTPUT_SECURE_BOOT"] = "3"

	trustSourceQuantifiers2, _ := createTrustSourceQuantifiers(params)

	trustSourceQuantifiers2[0].Quantifier(param)
	trustSourceQuantifiers2[1].Quantifier(param)

}

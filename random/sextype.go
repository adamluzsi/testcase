package random

import (
	"go.llib.dev/testcase/internal"
	"go.llib.dev/testcase/random/sextype"
)

func randomSexType(random *Random) internal.SexType {
	if random.Bool() {
		return sextype.Male
	} else {
		return sextype.Female
	}
}

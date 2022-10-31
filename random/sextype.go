package random

import (
	"github.com/adamluzsi/testcase/internal"
	"github.com/adamluzsi/testcase/random/sextype"
)

func randomSexType(random *Random) internal.SexType {
	if random.Bool() {
		return sextype.Male
	} else {
		return sextype.Female
	}
}

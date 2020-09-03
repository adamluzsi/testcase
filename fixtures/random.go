package fixtures

import (
	"math/rand"
	"time"

	"github.com/adamluzsi/testcase/fixtures/random"
)

var Random = random.NewRandom(rand.NewSource(time.Now().Unix()))
var SecureRandom = random.NewRandom(random.CryptoSeed{})

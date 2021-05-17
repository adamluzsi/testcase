package fixtures

import (
	"github.com/adamluzsi/testcase/random"
	"math/rand"
	"time"
)

var Random = random.NewRandom(rand.NewSource(time.Now().Unix()))
var SecureRandom = random.NewRandom(random.CryptoSeed{})

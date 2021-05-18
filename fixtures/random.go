package fixtures

import (
	"math/rand"
	"time"

	"github.com/adamluzsi/testcase/random"
)

var Random = random.New(rand.NewSource(time.Now().Unix()))
var SecureRandom = random.New(random.CryptoSeed{})

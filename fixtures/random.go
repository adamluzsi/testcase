package fixtures

import (
	"github.com/adamluzsi/testcase/random"
	"math/rand"
	"time"
)

var Random = random.New(rand.NewSource(time.Now().Unix()))
var SecureRandom = random.New(random.CryptoSeed{})

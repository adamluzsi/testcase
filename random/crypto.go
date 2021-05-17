package random

import (
	"crypto/rand"
	"math/big"
)

type CryptoSeed struct{}

func (c CryptoSeed) Uint64() uint64 {
	// MaxInt gives the current system's maximum integer
	// ^ means invert bits in the expression so if: uint(0) == 0000...0000 (exactly 32 or 64 zero bits depending on build target architecture)
	// then ^unit(0) == 1111...1111 which gives us the maximum value for the unsigned integer (all ones).
	// Then we need to shift integers to the right since the first bit store the sign in an int,
	// which gives us ^uint(0) >> 1 == 0111...1111.
	const MaxInt = int(^uint(0) >> 1)
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(MaxInt)))
	if err != nil {
		panic(err)
	}
	return nBig.Uint64()
}

// Int63 returns a non-negative pseudo-random 63-bit integer as an int64.
func (c CryptoSeed) Int63() int64 {
	const (
		rngMax  = 1 << 63
		rngMask = rngMax - 1
	)
	return int64(c.Uint64() & rngMask)
}

// Seed should use the provided seed value to initialize the generator to a deterministic state,
// but in CryptoSeed, the value is ignored.
func (c CryptoSeed) Seed(seed int64) {}

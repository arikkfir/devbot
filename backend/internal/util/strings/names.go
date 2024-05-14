package strings

import (
	"math/rand"

	"github.com/lucasepe/codename"
	"github.com/secureworks/errors"
)

var (
	rng *rand.Rand
)

func init() {
	if defaultRNG, err := codename.DefaultRNG(); err != nil {
		panic(errors.New("failed to initialize codename RNG"))
	} else {
		rng = defaultRNG
	}
}

func Name() string {
	return codename.Generate(rng, 0)
}

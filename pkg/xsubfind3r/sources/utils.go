package sources

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func PickRandom[T any](v []T) (picked T, err error) {
	length := len(v)

	if length == 0 {
		return
	}

	// Generate a cryptographically secure random index
	max := big.NewInt(int64(length))

	var indexBig *big.Int

	indexBig, err = rand.Int(rand.Reader, max)
	if err != nil {
		err = fmt.Errorf("failed to generate random index: %v", err)

		return
	}

	index := indexBig.Int64()

	// Return the element at the random index
	picked = v[index]

	return
}

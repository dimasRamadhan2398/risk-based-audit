package reference

import (
	_ "embed"
	"sync"
)

//go:embed data/regions.json
var regionsJSON []byte

//go:embed data/countries.json
var countriesJSON []byte

var (
	defaultOnce sync.Once
	defaultSet  *Dataset
	defaultErr  error
)

// Default returns the embedded dataset, parsed and validated on first use.
// serve calls it at start-up so invalid data stops the service there.
func Default() (*Dataset, error) {
	defaultOnce.Do(func() {
		defaultSet, defaultErr = Load(regionsJSON, countriesJSON)
	})
	return defaultSet, defaultErr
}

// MustDefault is Default for code paths that run after start-up validation.
// It panics only if the embedded data is invalid, which the package tests and
// the start-up check both rule out.
func MustDefault() *Dataset {
	d, err := Default()
	if err != nil {
		panic(err)
	}
	return d
}

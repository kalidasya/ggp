package gp

import (
	"golang.org/x/exp/slices"
	"math/rand"
)

func SelRandom(individuals []Individual, k int, r *rand.Rand) []Individual {
	chosen := make([]Individual, k)
	for i := range chosen {
		chosen[i] = individuals[r.Intn(len(individuals))].Copy()
	}
	return chosen
}

// TODO FitnessMaxFunc should be a parameter
func SelTournament(individuals []Individual, k, tournsize int, r *rand.Rand) []Individual {
	if k > len(individuals) {
		k = len(individuals)
	}
	chosen := make([]Individual, k)
	for i := 0; i < k; i++ {
		chosen[i] = slices.MaxFunc(SelRandom(individuals, tournsize, r), FitnessMaxFunc)
		// TODO see why we prefer one primitive here out of the others
	}
	return chosen
}

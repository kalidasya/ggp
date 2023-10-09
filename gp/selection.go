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
    // todo stats about what kind of individuals we chose here
		chosen[i] = slices.MaxFunc(SelRandom(individuals, tournsize, r), FitnessMaxFunc)
	}
	return chosen
}

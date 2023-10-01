package gp

import (
	"golang.org/x/exp/slices"
	"math/rand"
)

func SelRandom(individuals []Individual, k int, r *rand.Rand) []Individual {
	perm := r.Perm(len(individuals))
	chosen := make([]Individual, k)
	for i, randIndex := range perm {
		if i >= k {
			break
		}
		chosen[i] = individuals[randIndex].Copy()
	}
	return chosen
}

func SelTournament(individuals []Individual, k int, tournsize int, r *rand.Rand) []Individual {
	if k > len(individuals) {
		k = len(individuals)
	}
	chosen := make([]Individual, k)
	for i := 0; i < k; i++ {
		aspirants := SelRandom(individuals, tournsize, r)
		sel := slices.MaxFunc(aspirants, FitnessMaxFunc)
		chosen[i] = sel.Copy()
	}
	return chosen
}

package gp

import (
	"fmt"
	"golang.org/x/exp/slices"
	"math/rand"
)

// TODO make mutator and CX function a parameter
func VarAnd(offs []Individual, ps *PrimitiveSet, cxFunc CrossOver, mutFunc Mutator, cxpb, mutpb float32, r *rand.Rand) {
	for i := 1; i < len(offs); i += 2 {
		if rand.Float32() < cxpb {
			tree1, tree2 := cxFunc(*offs[i-1].Tree(), *offs[i].Tree(), r, 0)
			offs[i-1].Tree().ReplaceNodes(tree1.Nodes())
			offs[i].Tree().ReplaceNodes(tree2.Nodes())
			offs[i-1].Fitness().DelValues()
			offs[i].Fitness().DelValues()
		}
	}
	for i := 0; i < len(offs); i++ {
		if rand.Float32() < mutpb {
			offs[i].Tree().ReplaceNodes(
				mutFunc(offs[i].Tree()).Nodes(),
			)
			offs[i].Fitness().DelValues()
		}
	}
}

type AlgorithmSettings struct {
	NumGen               int
	TournamentSize       int
	SelectionSize        int
	MutationProbability  float32
	CrossoverProbability float32
	CrossOverFunc        CrossOver
	MutatorFunc          Mutator
}

func EaSimple(inds []Individual, ps *PrimitiveSet, evalFunction func(Individual), setting AlgorithmSettings, r *rand.Rand) []Individual {
	for gen := 0; gen < setting.NumGen; gen++ {
		fmt.Printf("------------------------------------------------------------------- (%d) %d\n", gen+1, len(inds))
		offsprings := SelTournament(inds, setting.SelectionSize, setting.TournamentSize, r)
		// TODO pass on settings?
		VarAnd(offsprings, ps, setting.CrossOverFunc, setting.MutatorFunc, setting.CrossoverProbability, setting.MutationProbability, r)
		for i := range offsprings {
			if !offsprings[i].Fitness().Valid() {
				evalFunction(offsprings[i])
			}
		}
		best := slices.MaxFunc(inds, FitnessMaxFunc)
		fmt.Printf("Best in gen: %s\n", best.Fitness().String())
		inds = offsprings
	}

	best := slices.MaxFunc(inds, FitnessMaxFunc)
	fmt.Printf("Final: %s %s\n", best.Fitness().String(), best.Tree().String())
	return inds
}

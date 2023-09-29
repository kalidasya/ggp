package gp

import (
	"fmt"
	"golang.org/x/exp/slices"
	"math/rand"
	"reflect"
)

func VarAnd(offs []Individual, ps *PrimitiveSet, cxpb, mutpb float32, r *rand.Rand) []Individual {
	for i := 1; i < len(offs); i += 2 {
		if rand.Float32() < cxpb {
			tree1, tree2 := CXOnePoint(*offs[i-1].Tree(), *offs[i].Tree(), r, 0)
			offs[i-1].Tree().ReplaceNodes(tree1.Nodes())
			offs[i].Tree().ReplaceNodes(tree2.Nodes())
			offs[i-1].Fitness().DelValues()
			offs[i].Fitness().DelValues()
		}
	}
	for i := 0; i < len(offs); i++ {
		uniformMutator := NewUniformMutator(ps, func(ps *PrimitiveSet, type_ reflect.Kind) []Node {
			return GenerateTree(ps, 0, 2, GenFull, type_, r).Nodes()
		}, r)
		if rand.Float32() < mutpb {
			offs[i].Tree().ReplaceNodes(
				StaticMutatorLimiter(uniformMutator.Mutate, 9999)(offs[i].Tree()).Nodes(),
			)
			offs[i].Fitness().DelValues()
		}
	}
	return offs
}

func EaSimple(inds []Individual, ps *PrimitiveSet, evalFunction func(Individual), r *rand.Rand) {
	for gen := 0; gen < 40; gen++ {
		fmt.Printf("------------------------------------------------------------------- (%d) %d\n", gen+1, len(inds))
		offsprings := SelTournament(inds, len(inds), 7, r)
		offsprings = VarAnd(offsprings, ps, 0.5, 0.2, r)
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
	fmt.Printf("Final: %s %s", best.Fitness().String(), best.Tree().String())
}

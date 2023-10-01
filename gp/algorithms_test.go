package gp

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func generateInds(amount int, initialFitness, initialWeight float32, ps *PrimitiveSet, r *rand.Rand) []Individual {
	inds := []Individual{}
	for i := 0; i < amount; i++ {
		fit, _ := NewFitness([]float32{initialWeight})
		fit.SetValues([]float32{initialFitness})
		inds = append(inds, &IndividualImpl{
			tree:    GenerateTree(ps, 1, 2, GenFull, ps.RetType, r),
			fitness: fit,
		})
	}
	return inds
}

func TestVarAndNoChange(t *testing.T) {
	r := rand.New(rand.NewSource(14))
	ps := getPrimitiveSet()

	inds := generateInds(10, 1, 2, ps, r)
	VarAnd(inds, ps, 0, 0, r)

	for i := range inds {
		assert.True(t, inds[i].Fitness().Valid())
	}
}

func TestVarAndMutationOnly(t *testing.T) {
	r := rand.New(rand.NewSource(33))
	ps := getPrimitiveSet()

	inds := generateInds(10, 1, 2, ps, r)
	VarAnd(inds, ps, 0, 1, r)

	for i := range inds {
		assert.False(t, inds[i].Fitness().Valid())
	}
	inds[0].Fitness().SetValues([]float32{4})
	for i := range inds {
		if i == 0 {
			assert.Equal(t, []float32{8}, inds[i].Fitness().GetWValues())
		} else {
			assert.False(t, inds[i].Fitness().Valid())
		}
	}
}

func TestVarAndCXOnly(t *testing.T) {
	r := rand.New(rand.NewSource(12))
	ps := getPrimitiveSet()

	inds := generateInds(10, 1, 2, ps, r)
	VarAnd(inds, ps, 1, 0, r)

	for i := range inds {
		assert.False(t, inds[i].Fitness().Valid())
	}
	inds[0].Fitness().SetValues([]float32{4})
	for i := range inds {
		if i == 0 {
			assert.Equal(t, []float32{8}, inds[i].Fitness().GetWValues())
		} else {
			assert.False(t, inds[i].Fitness().Valid())
		}
	}
}

func TestVarAndCXAndMut(t *testing.T) {
	r := rand.New(rand.NewSource(20))
	ps := getPrimitiveSet()

	inds := generateInds(10, 1, 2, ps, r)
	VarAnd(inds, ps, 1, 1, r)

	for i := range inds {
		assert.False(t, inds[i].Fitness().Valid())
	}
	inds[0].Fitness().SetValues([]float32{4})
	for i := range inds {
		if i == 0 {
			assert.Equal(t, []float32{8}, inds[i].Fitness().GetWValues())
		} else {
			assert.False(t, inds[i].Fitness().Valid())
		}
	}
}

func TestEaSimple(t *testing.T) {
	r := rand.New(rand.NewSource(20))
	ps := getPrimitiveSet()

	inds := generateInds(10, 1, 2, ps, r)
	evalFunc := func(ind Individual) {
		// resetting all fitness
		ind.Fitness().SetValues([]float32{float32(len(ind.Tree().Nodes())) / 2.0})
	}
	setting := AlgorithmSettings{
		NumGen:               5,
		MutationProbability:  1,
		CrossoverProbability: 1,
		TournamentSize:       6,
		SelectionSize:        len(inds),
	}
	inds = EaSimple(inds, ps, evalFunc, setting, r)

	for i := range inds {
		assert.True(t, inds[i].Fitness().Valid())
		assert.Greater(t, inds[i].Fitness().GetWValues()[0], float32(2))
		assert.Less(t, inds[i].Fitness().GetWValues()[0], float32(12))
	}

}

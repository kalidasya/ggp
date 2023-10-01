package gp

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"reflect"
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

func getMutator(ps *PrimitiveSet, r *rand.Rand) Mutator {
	return StaticMutatorLimiter(NewUniformMutator(ps, func(ps *PrimitiveSet, type_ reflect.Kind) []Node {
		return GenerateTree(ps, 0, 2, GenGrow, type_, r).Nodes()
	}, r).Mutate, 17)
}

func getCrossOver() CrossOver {
	return StaticCrossOverLimiter(CXOnePoint, 17)
}

func TestVarAndNoChange(t *testing.T) {
	r := rand.New(rand.NewSource(14))
	ps := getPrimitiveSet()

	inds := generateInds(10, 1, 2, ps, r)
	VarAnd(inds, ps, getCrossOver(), getMutator(ps, r), 0, 0, r)

	for i := range inds {
		assert.True(t, inds[i].Fitness().Valid())
	}
}

func TestVarAndMutationOnly(t *testing.T) {
	r := rand.New(rand.NewSource(33))
	ps := getPrimitiveSet()

	inds := generateInds(10, 1, 2, ps, r)
	VarAnd(inds, ps, getCrossOver(), getMutator(ps, r), 0, 1, r)

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
	VarAnd(inds, ps, getCrossOver(), getMutator(ps, r), 1, 0, r)

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
	VarAnd(inds, ps, getCrossOver(), getMutator(ps, r), 1, 1, r)

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
		NumGen:               10,
		MutationProbability:  0.5,
		CrossoverProbability: 0.5,
		TournamentSize:       4,
		SelectionSize:        len(inds),
		CrossOverFunc:        getCrossOver(),
		MutatorFunc:          getMutator(ps, r),
	}
	inds = EaSimple(inds, ps, evalFunc, setting, r)

	for i := range inds {
		assert.True(t, inds[i].Fitness().Valid())
		assert.Greater(t, inds[i].Fitness().GetWValues()[0], float32(2))
		assert.Less(t, inds[i].Fitness().GetWValues()[0], float32(12))
	}

}

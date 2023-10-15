package gp

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

type IndividualImpl struct {
	tree    *PrimitiveTree
	fitness *Fitness
}

func (a *IndividualImpl) Fitness() *Fitness {
	return a.fitness
}

func (a *IndividualImpl) Tree() *PrimitiveTree {
	return a.tree
}

func (a *IndividualImpl) Copy() Individual {
	fit, err := NewFitness(a.Fitness().GetWeights())
	if err != nil {
		panic(err)
	}
	return &IndividualImpl{
		tree:    NewPrimitiveTree(a.Tree().Nodes()),
		fitness: fit,
	}
}

func TestSelRandom(t *testing.T) {
	r := rand.New(rand.NewSource(17))
	inds := []Individual{}

	for i := 0; i < 10; i++ {
		fit, _ := NewFitness([]float32{2.0})
		inds = append(inds, &IndividualImpl{
			tree:    NewPrimitiveTree(getValidNodes()),
			fitness: fit,
		})
	}

	result := SelRandom(inds, 5, r)
	assert.Len(t, result, 5)
	assert.NotEqual(t, inds, result)

	for i := 0; i < len(inds)-1; i++ {
		inds[i].Fitness().SetValues([]float32{4})
		assert.Equal(t, []float32{8}, inds[i].Fitness().GetWValues())
		assert.False(t, inds[i+1].Fitness().Valid())
	}
}

func TestSelTournament(t *testing.T) {
	r := rand.New(rand.NewSource(17))
	inds := []Individual{}
	for i := 0; i < 10; i++ {
		fit, _ := NewFitness([]float32{2.0})
		inds = append(inds, &IndividualImpl{
			tree:    NewPrimitiveTree(getValidNodes()),
			fitness: fit,
		})
		if i < 5 {
			inds[len(inds)-1].Fitness().SetValues([]float32{5})
		}
	}

	for i := 0; i < 5; i++ {
		assert.Equal(t, []float32{10}, inds[i].Fitness().GetWValues())
	}
	for i := 5; i < len(inds); i++ {
		assert.Empty(t, inds[i].Fitness().GetWValues())
	}

	result := SelTournament(inds, 8, 5, FitnessMaxFunc, r)
	assert.Len(t, result, 8)
	assert.NotEqual(t, inds, result)

	for i := 0; i < len(inds)-1; i++ {
		inds[i].Fitness().SetValues([]float32{4})
		assert.Equal(t, []float32{8}, inds[i].Fitness().GetWValues())
		assert.NotEqual(t, []float32{8}, inds[i+1].Fitness().GetWValues())
	}

}

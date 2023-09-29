package gp

import (
	"fmt"
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
	assert.Equal(t, []Individual{inds[8], inds[7], inds[1], inds[0], inds[2]}, result)
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
		fmt.Printf("%d %d\n", i, inds[len(inds)-1])
	}

	result := SelTournament(inds, 8, 5, r)
	for i := 0; i < len(result); i++ {
		fmt.Printf("%d %d\n", i, result[i])
	}
	assert.Len(t, result, 8)
	assert.Equal(t, []Individual{inds[8], inds[0], inds[0], inds[5],
		inds[1], inds[8], inds[4], inds[7]}, result)
}

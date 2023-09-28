package gp

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var func1 PrimitiveFunc = func(a ...PrimitiveArgs) PrimitiveArgs {
	return len(a[1].(string)) * a[0].(int)
}
var func2 PrimitiveFunc = func(a ...PrimitiveArgs) PrimitiveArgs {
	return strings.Repeat(a[0].(string), a[1].(int))
}

var prim1 = NewPrimitive("prim1", func1, []reflect.Kind{reflect.Int, reflect.String}, reflect.Int)
var prim2 = NewPrimitive("prim2", func2, []reflect.Kind{reflect.String, reflect.Int}, reflect.String)
var term1 = NewTerminal("term1", reflect.Int, 4)
var term2 = NewTerminal("term2", reflect.String, "hello")

func getValidNodes() []Node {
	return []Node{
		prim1, term1, prim2, term2, term1,
	}
}

func getValidNodes2() []Node {
	return []Node{
		prim2, term2, prim1, term1, term2,
	}
}

func getPrimitiveSet() *PrimitiveSet {
	ps := NewPrimitiveSet([]reflect.Kind{}, reflect.Int)
	ps.AddPrimitive(prim1)
	ps.AddPrimitive(prim2)
	ps.AddTerminal(term1)
	ps.AddTerminal(term2)

	return ps
}

func TestPrimitiveTreeString(t *testing.T) {
	tree := NewPrimitiveTree(getValidNodes())
	assert.Equal(t, `prim1(4, prim2("hello", 4))`, tree.String())
}

func TestPrimitiveTreeCompile(t *testing.T) {
	tree := NewPrimitiveTree(getValidNodes())
	assert.Equal(t, 5*4*4, tree.Compile().(int))
}

func TestPrimitiveTreeRoot(t *testing.T) {
	nodes := getValidNodes()
	tree := NewPrimitiveTree(nodes)
	assert.Equal(t, nodes[0], tree.Root())
}

func TestPrimitiveTreeHeight(t *testing.T) {
	tree := NewPrimitiveTree(getValidNodes())
	assert.Equal(t, 2, tree.Height())
}

func TestSearchSubtree(t *testing.T) {
	type TestCase struct {
		index         int
		expectedStart int
		expectedEnd   int
	}
	nodes := getValidNodes()
	tree := NewPrimitiveTree(nodes)
	for _, scenario := range []TestCase{
		{
			index:         0,
			expectedStart: 0,
			expectedEnd:   len(nodes),
		},
		{
			index:         1,
			expectedStart: 1,
			expectedEnd:   2,
		},
		{
			index:         2,
			expectedStart: 2,
			expectedEnd:   5,
		},
	} {
		actualStart, actualEnd := tree.SearchSubtree(scenario.index)
		assert.Equal(t, scenario.expectedStart, actualStart)
		assert.Equal(t, scenario.expectedEnd, actualEnd)
	}
}

func TestGenerateTree(t *testing.T) {
	ps := getPrimitiveSet()

	r := rand.New(rand.NewSource(17))
	tree1 := GenerateTree(ps, 3, 5, GenFull, ps.RetType, r)
	r = rand.New(rand.NewSource(17))
	tree1Again := GenerateTree(ps, 3, 5, GenFull, ps.RetType, r)
	tree2 := GenerateTree(ps, 3, 5, GenGrow, ps.RetType, r)

	assert.Equal(t, tree1.Nodes(), tree1Again.Nodes())
	assert.NotEqual(t, tree1.Nodes(), tree2.Nodes())
	assert.GreaterOrEqual(t, tree1.Height(), 3)
	assert.LessOrEqual(t, tree1.Height(), 5)
	assert.GreaterOrEqual(t, tree2.Height(), 3)
	assert.LessOrEqual(t, tree2.Height(), 5)
	assert.NotPanics(t, func() { tree1.Compile() })
	assert.NotPanics(t, func() { tree2.Compile() })
}

func TestCompileWithArguments(t *testing.T) {
	r := rand.New(rand.NewSource(1001))

	ps := NewPrimitiveSet([]reflect.Kind{reflect.Int, reflect.String, reflect.Int}, reflect.Int)
	ps.AddPrimitive(prim1)
	ps.AddPrimitive(prim2)

	assert.Equal(t, 2, len(ps.Terminals[reflect.Int]))
	assert.Equal(t, 1, len(ps.Terminals[reflect.String]))
	tree := GenerateTree(ps, 2, 3, GenFull, ps.RetType, r)
	assert.NotPanics(t, func() { tree.Compile(4, "aloha", 2) })
	assert.Equal(t, 0, tree.Compile(1, "", 1))
	// panic if not enough argument
	assert.Panics(t, func() { tree.Compile(4, "aloha") })
	// panic if type mismtach
	assert.Panics(t, func() { tree.Compile(true, "aloha", 1) })
	// no panic if extra arguments
	assert.NotPanics(t, func() { tree.Compile(1, "aloha", 1, 12) })

}

func TestUniformMutator(t *testing.T) {
	r := rand.New(rand.NewSource(4853))
	tree := NewPrimitiveTree(getValidNodes())
	ps := getPrimitiveSet()
	origLen := len(tree.Nodes())
	beforeMut := fmt.Sprintf("%s", tree)
	uniformMutator := NewUniformMutator(ps, func(ps *PrimitiveSet, type_ reflect.Kind) []Node {
		return GenerateTree(ps, 1, 2, GenGrow, type_, r).Nodes()
	}, r)
	tree = uniformMutator.Mutate(tree)
	assert.Len(t, tree.Nodes(), origLen+2) // adds 3 nodes and removes one
	assert.NotEqual(t, beforeMut, fmt.Sprintf("%s", tree))
	assert.Equal(t, `prim1(4, prim2(prim2("hello", 4), 4))`, fmt.Sprintf("%s", tree))
	assert.NotPanics(t, func() { tree.Compile() })
}

func TestCXOnePointestCXOnePoint(t *testing.T) {
	tree1 := &PrimitiveTree{
		stack: []Node{prim1, prim1, prim1, term1, term2, prim2, term2, term1, prim2, term2, prim1, term1, term2}}
	assert.NotPanics(t, func() { tree1.Compile() })
	tree2 := &PrimitiveTree{
		stack: []Node{prim2, prim2, term2, term1, prim1, term1, term2},
	}
	assert.NotPanics(t, func() { tree2.Compile() })

	r := rand.New(rand.NewSource(715))
	tree1, tree2 = CXOnePoint(tree1, tree2, r, 0) // node at index 3 in tree1 will be replaced with node index 4:7 in tree2
	assert.Equal(t, []Node{prim1, prim1, prim1, prim1, term1, term2, term2, prim2, term2, term1, prim2, term2, prim1, term1, term2}, tree1.stack)
	assert.Equal(t, []Node{prim2, prim2, term2, term1, term1}, tree2.stack)
	assert.NotPanics(t, func() { tree1.Compile() })
	assert.NotPanics(t, func() { tree2.Compile() })
}

func TestStaticCrossOverLimiter(t *testing.T) {
	ps := getPrimitiveSet()
	r := rand.New(rand.NewSource(444))
	tree1 := GenerateTree(ps, 3, 4, GenFull, ps.RetType, r)
	tree2 := GenerateTree(ps, 3, 4, GenFull, ps.RetType, r)
	for i := 0; i < 10; i++ {
		tree1, tree2 = StaticCrossOverLimiter(CXOnePoint, 17)(tree1, tree2, r, 0)
		assert.Less(t, len(tree1.Nodes()), 18)
		assert.Less(t, len(tree2.Nodes()), 18)
	}
}

func TestStaticMutatorLimiter(t *testing.T) {
	ps := getPrimitiveSet()
	r := rand.New(rand.NewSource(444))
	tree := GenerateTree(ps, 3, 4, GenFull, ps.RetType, r)
	uniformMutator := NewUniformMutator(ps, func(ps *PrimitiveSet, type_ reflect.Kind) []Node {
		return GenerateTree(ps, 2, 3, GenGrow, type_, r).Nodes()
	}, r)
	for i := 0; i < 10; i++ {
		tree = StaticMutatorLimiter(uniformMutator.Mutate, 17)(tree)
		assert.Less(t, len(tree.Nodes()), 18)
	}
}

package gp

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var func1 PrimitiveFunc = func(a ...interface{}) interface{} {
	return len(a[1].(string)) * a[0].(int)
}
var func2 PrimitiveFunc = func(a ...interface{}) interface{} {
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
	ps := NewPrimitiveSet([]reflect.Kind{reflect.Int, reflect.String}, reflect.Int)
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

}

func TestMutUniform(t *testing.T) {
	r := rand.New(rand.NewSource(4853))
	tree := NewPrimitiveTree(getValidNodes())
	ps := getPrimitiveSet()
	origLen := len(tree.Nodes())
	beforeMut := fmt.Sprintf("%s", tree)
	MutUniform(tree, func(ps *PrimitiveSet, type_ reflect.Kind) []Node {
		return GenerateTree(ps, 1, 2, GenGrow, type_, r).Nodes()
	}, ps, r)
	assert.Len(t, tree.Nodes(), origLen+2) // adds 3 nodes and removes one
	assert.NotEqual(t, beforeMut, fmt.Sprintf("%s", tree))
	assert.Equal(t, `prim1(4, prim2(prim2("hello", 4), 4))`, fmt.Sprintf("%s", tree))
	assert.NotNil(t, tree.Compile())
}

/*
var prim1 = NewPrimitive("func1", func1, []reflect.Kind{reflect.Int, reflect.String}, reflect.Int)
var prim2 = NewPrimitive("func2", func2, []reflect.Kind{reflect.String, reflect.Int}, reflect.String)
var term1 = NewTerminal("term1", reflect.Int, 4)
var term2 = NewTerminal("term2", reflect.String, "hello")
*/

func TestCXOnePoint(t *testing.T) {
	tree1 := &PrimitiveTree{
		stack: []Node{prim1, prim1, prim1, term1, term2, prim2, term2, term1, prim2, term2, term1, prim1, term1, term2}}

	PrintNodes(tree1.Nodes())
	fmt.Printf("%s\n", tree1)
	assert.NotNil(t, tree1.Compile())
	tree2 := &PrimitiveTree{
		stack: []Node{prim2, prim2, term2, term1, prim1, term1, term2},
	}
	assert.NotNil(t, tree2.Compile())

	fmt.Println(tree1)
	fmt.Println(tree2)
	r := rand.New(rand.NewSource(715))

	fmt.Println("------------------")
	CXOnePoint(tree1, tree2, r) // node at index 9 in tree 1 will be replaced with node index 2 in tree 2
	fmt.Println(tree1)
	fmt.Println(tree2)
	assert.Equal(t, `prim1(prim1(prim2("hello", 4), prim2("hello", 4)), prim1(prim2("hello", prim1(4, "hello")), prim1(4, "hello")))`, fmt.Sprintf("%s", tree1))
	assert.Equal(t, `prim2(prim2("hello", 4), 4)`, fmt.Sprintf("%s", tree2))
	tree1.Compile()
	tree2.Compile()
}

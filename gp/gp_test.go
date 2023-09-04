package gp

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type FakeRandom struct {
	ints    []int
	floats  []float32
	cint    int
	cfloats int
}

func (r *FakeRandom) Intn(n int) int {
	ret := r.ints[r.cint]
	r.cint++
	fmt.Printf("asked %d. rand int", r.cint)
	return ret
}

func (r *FakeRandom) Float32() float32 {
	ret := r.floats[r.cfloats]
	r.cfloats++
	return ret
}

var func1 PrimitiveFunc = func(a ...interface{}) interface{} {
	return len(a[1].(string)) * a[0].(int)
}
var func2 PrimitiveFunc = func(a ...interface{}) interface{} {
	return strings.Repeat(a[0].(string), a[1].(int))
}

var prim1 = NewPrimitive("func1", func1, []reflect.Kind{reflect.Int, reflect.String}, reflect.Int)
var prim2 = NewPrimitive("func2", func2, []reflect.Kind{reflect.String, reflect.Int}, reflect.String)
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
	assert.Equal(t, `func1(4, func2("hello", 4))`, tree.String())
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

func TestMutUniform(t *testing.T) {
	tree := NewPrimitiveTree(getValidNodes())
	ps := getPrimitiveSet()
	r := &FakeRandom{ints: []int{2, 1, 1, 0, 0, 0, 0, 0, 0}}
	origLen := len(tree.Nodes())
	MutUniform(tree, func(ps *PrimitiveSet, type_ reflect.Kind) []Node {
		return GenerateTree(ps, 1, 2, GenGrow, type_, r).Nodes()
	}, ps, r)
	assert.Len(t, tree.Nodes(), origLen+2) // adds 3 nodes and removes one
	assert.Equal(t, `func1(4, func2(func1(4, "hello"), 4))`, fmt.Sprintf("%s", tree))
	tree.Compile()
}

func TestCXOnePoint(t *testing.T) {
	tree1 := &PrimitiveTree{
		stack: []Node{prim1, prim1, prim2, term2, term1, prim2, term2, term1, prim1, prim2, term2, term1, prim1, term1, term2},
	}
	tree2 := &PrimitiveTree{
		stack: []Node{prim2, prim2, term2, term1, prim1, term1, term2},
	}

	fmt.Println(tree1)
	fmt.Println(tree2)
	r := &FakeRandom{ints: []int{1, 4, 1}}

	fmt.Println("------------------")
	CXOnePoint(tree1, tree2, r) // node at index 9 in tree 1 will be replaced with node index 2 in tree 2
	tree1.Compile()
	tree2.Compile()
	assert.Equal(t, `func1(func1(func2("hello", 4), func2("hello", 4)), func1("hello", func1(4, "hello")))`, fmt.Sprintf("%s", tree1))
	assert.Equal(t, `func1(func1(func2("hello", 4), func2("hello", 4)), func1("hello", func1(4, "hello")))`, fmt.Sprintf("%s", tree1))
	fmt.Println(tree1)
	fmt.Println(tree2)
}

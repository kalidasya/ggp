package gp

import (
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

var prim1 = NewPrimitive("func1", func1, []reflect.Kind{reflect.Int, reflect.String}, reflect.Int)
var prim2 = NewPrimitive("func2", func2, []reflect.Kind{reflect.String, reflect.Int}, reflect.String)
var term1 = NewTerminal("term1", reflect.Int, 4)
var term2 = NewTerminal("term2", reflect.String, "hello")

func getValidNodes() []Node {
	return []Node{
		prim1, term1, prim2, term2, term1,
	}
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

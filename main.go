package main

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
)

// todo use slices when go is 1.20
func ReverseSlice[T comparable](s []T) []T {
	var r []T
	for i := len(s) - 1; i >= 0; i-- {
		r = append(r, s[i])
	}
	return r
}

type PrimitiveFunc func(...interface{}) interface{}

type GenCondition func(int, int, int, int, *PrimitiveSet) bool

var GenGrow GenCondition = func(height int, depth int, min int, max int, ps *PrimitiveSet) bool {
	return depth == height || (depth >= min && rand.Float32() < ps.TerminalRatio())
}

var GenFull GenCondition = func(height int, depth int, _min int, _max int, _ps *PrimitiveSet) bool {
	return depth == height
}

func Max(a int, b int) int {
	if a >= b {
		return a
	}
	return b
}

type PrimitiveTree struct {
	stack []Node // either primitive or terminal
}

type NodeString struct {
	node Node
	str  []string
}

func (pt *PrimitiveTree) String() string {
	s := ""
	var stack []NodeString
	for _, node := range pt.stack {
		stack = append(stack, NodeString{node, []string{}})
		for len(stack[len(stack)-1].str) == stack[len(stack)-1].node.Arity() {
			var n NodeString
			stack, n = Pop(stack)
			s = n.node.Str(n.str)
			if len(stack) == 0 {
				break
			}
			stack[len(stack)-1].str = append(stack[len(stack)-1].str, s)
		}
	}
	return s
}

func (pt *PrimitiveTree) Root() interface{} {
	return pt.stack[0]
}

func Append(s []int, times int, value int) []int {
	for i := 0; i < times; i++ {
		s = append(s, value)
	}
	return s
}

func (pt *PrimitiveTree) Height() int {
	stack := []int{0}
	max_depth := 0
	var depth int
	for _, elem := range pt.stack {
		stack, depth = Pop(stack)
		max_depth = Max(max_depth, depth)
		stack = Append(stack, elem.Arity(), depth+1)
	}
	return max_depth
}

func (pt *PrimitiveTree) SearchSubtree(begin int) (int, int) {
	end := begin + 1
	total := pt.stack[begin].Arity()
	for total > 0 {
		total += pt.stack[end].Arity() - 1
		end++
	}
	return begin, end
}

func NewPrimitiveTree(stack []Node) *PrimitiveTree {
	return &PrimitiveTree{
		stack: stack,
	}
}

type Node interface {
	Arity() int
	Name() string
	Eval() (interface{}, error)
	Str([]string) string
}

type Terminal struct {
	name    string
	retType reflect.Kind
	value   interface{}
}

func (t *Terminal) Arity() int {
	return 0
}

func (t *Terminal) Name() string {
	return t.name
}

func (t *Terminal) Eval() (interface{}, error) {
	return t.value, nil
}

func (t *Terminal) Str(_ []string) string {
	return fmt.Sprintf("%v", t.value)
}

var _ Node = new(Terminal)

func NewTerminal(name string, retType reflect.Kind, value interface{}) *Terminal {
	return &Terminal{
		name:    name,
		retType: retType,
		value:   value,
	}
}

type Primitive struct {
	name     string
	function PrimitiveFunc
	arity    int
	argTypes []reflect.Kind
	retType  reflect.Kind
	args     []interface{}
}

func (p *Primitive) Arity() int {
	return p.arity
}

func (p *Primitive) Name() string {
	return p.name
}

func (p *Primitive) Equals(o Primitive) bool {
	if p.arity != o.arity {
		return false
	}
	for i := 0; i < len(p.argTypes); i++ {
		if p.argTypes[i] != o.argTypes[i] {
			return false
		}
	}
	if p.retType != o.retType {
		return false
	}
	return true
}

func (p *Primitive) Eval() (interface{}, error) {
	if len(p.argTypes) > len(p.args) {
		return nil, errors.New("not enough arguments")
	}
	if len(p.argTypes) < len(p.args) {
		return nil, errors.New("too many arguments")
	}
	for i, arg := range p.args {
		if reflect.TypeOf(arg).Kind() != p.argTypes[i] {
			return nil, errors.New(fmt.Sprintf("invalid type for %dth argument", i))
		}

	}
	return p.function(p.args...), nil
}

func (p *Primitive) Str(args []string) string {
	return fmt.Sprintf("%s(%s)", p.Name(), strings.Join(args, ", "))
}

var _ Node = new(Primitive)

func NewPrimitive(name string, f PrimitiveFunc, argTypes []reflect.Kind, retType reflect.Kind) *Primitive {
	return &Primitive{
		name:     name,
		function: f,
		arity:    len(argTypes),
		argTypes: argTypes,
		retType:  retType,
	}
}

type PrimitiveSet struct {
	Primitives map[reflect.Kind][]*Primitive
	Terminals  map[reflect.Kind][]*Terminal
	InTypes    []reflect.Kind
	RetType    reflect.Kind
	Arity      int
}

func (ps *PrimitiveSet) AddPrimitive(p *Primitive) {
	for _, argType := range p.argTypes {
		val := ps.Primitives[argType]
		ps.Primitives[argType] = append(val, p)
	}
}

func (ps *PrimitiveSet) AddTerminal(t *Terminal) {
	val := ps.Terminals[t.retType]
	ps.Terminals[t.retType] = append(val, t)
}

func (ps *PrimitiveSet) TerminalRatio() float32 {
	return float32(len(ps.Terminals)) / float32(len(ps.Terminals)+len(ps.Primitives))
}

func NewPrimitiveSet(inTypes []reflect.Kind, retType reflect.Kind) *PrimitiveSet {
	return &PrimitiveSet{
		Primitives: make(map[reflect.Kind][]*Primitive),
		Terminals:  make(map[reflect.Kind][]*Terminal),
		RetType:    retType,
		InTypes:    inTypes,
		Arity:      len(inTypes),
	}
}

func Pop[T any](s []T) ([]T, T) {
	return s[:len(s)-1], s[len(s)-1]
}

type StackItem struct {
	i int
	t reflect.Kind
}

func GenerateTree(ps *PrimitiveSet, min int, max int, condition GenCondition) *PrimitiveTree {
	var expr []Node
	height := rand.Intn(max-min) + min

	stack := []StackItem{
		{i: 0, t: ps.RetType},
	}

	for len(stack) != 0 {
		var item StackItem
		stack, item = Pop(stack)
		depth := item.i
		realType := item.t
		if condition(height, depth, min, max, ps) {
			term := ps.Terminals[realType][rand.Intn(len(ps.Terminals[realType]))]
			if term == nil {
				panic("No terminal with type available")
			}
			expr = append(expr, term)
		} else {
			prim := ps.Primitives[realType][rand.Intn(len(ps.Primitives[realType]))]
			if prim == nil {
				panic("No primitive with type available")
			}
			expr = append(expr, prim)
			for i := len(prim.argTypes) - 1; i >= 0; i-- {
				arg := prim.argTypes[i]
				stack = append(stack, StackItem{i: depth + 1, t: arg})
			}
		}
	}
	return NewPrimitiveTree(expr)
}

func main() {
	var fone, ftwo PrimitiveFunc
	fone = func(a ...interface{}) interface{} {
		ret := 0
		for _, n := range a {
			ret += n.(int)
		}
		return ret
	}
	ftwo = func(a ...interface{}) interface{} {
		ret := 1
		for _, n := range a {
			ret *= n.(int)
		}
		return ret
	}
	p := NewPrimitive("fone", fone, []reflect.Kind{reflect.Int, reflect.Int}, reflect.Int)
	p2 := NewPrimitive("ftwo", ftwo, []reflect.Kind{reflect.Int, reflect.Int}, reflect.Int)
	t := NewTerminal("a", reflect.Int, 99)
	t2 := NewTerminal("b", reflect.Int, 12)
	// fmt.Println(p.Eval(8, 7))
	// fmt.Println(p2.Eval(3, 4))

	ps := NewPrimitiveSet([]reflect.Kind{reflect.Int, reflect.Int}, reflect.Int)
	ps.AddPrimitive(p)
	ps.AddPrimitive(p2)
	ps.AddTerminal(t)
	ps.AddTerminal(t2)
	ret := GenerateTree(ps, 3, 4, GenFull)
	fmt.Printf("tree depth: %d\n", ret.Height())
	fmt.Printf("tree: %s\n", ret)
	fmt.Println(ftwo(ftwo(ftwo(12, 99), fone(99, 99)), fone(ftwo(12, 99), fone(99, 99))))
}

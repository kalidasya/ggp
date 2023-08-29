package main

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
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
	stack []interface{}
}

func (pt *PrimitiveTree) Root() interface{} {
	return pt.stack[0]
}

func (pt *PrimitiveTree) Height() int {
	stack := []int{0}
	max_depth := 0
	for _, elem := range pt.stack.stack {
		depth, _ := stack.Pop()
		max_depth = Max(max_depth, depth)
		stack = append(stack)
		stack.extend([depth + 1]*elem.Arity)
	}
	return max_depth
}

type Primitive struct {
	function PrimitiveFunc
	arity    int
	argTypes []reflect.Kind
	retType  reflect.Kind
	args     []interface{}
}

type Terminal struct {
	Primitive
	value interface{}
}

func NewTerminal(retType reflect.Kind, value interface{}) *Terminal {
	return &Terminal{
		Primitive: Primitive{
			arity:   0,
			retType: retType,
		},
		value: value,
	}
}

func (t *Terminal) Eval() interface{} {
	return t.value
}

func NewPrimitive(f PrimitiveFunc, argTypes []reflect.Kind, retType reflect.Kind) *Primitive {
	return &Primitive{
		function: f,
		arity:    len(argTypes),
		argTypes: argTypes,
		retType:  retType,
	}
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

func (p *Primitive) Eval(args ...interface{}) (interface{}, error) {
	if len(p.argTypes) > len(args) {
		return nil, errors.New("not enough arguments")
	}
	if len(p.argTypes) < len(args) {
		return nil, errors.New("too many arguments")
	}
	for i, arg := range args {
		if reflect.TypeOf(arg).Kind() != p.argTypes[i] {
			return nil, errors.New(fmt.Sprintf("invalid type for %dth argument", i))
		}

	}
	return p.function(args...), nil
}

type PrimitiveSet struct {
	Primitives map[reflect.Kind][]*Primitive
	Terminals  map[reflect.Kind][]*Terminal
	InTypes    []reflect.Kind
	RetType    reflect.Kind
	Arity      int
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

func Pop[T any](s []T) ([]T, T) {
	item, stack := s[len(s)-1], s[:len(s)-1]
	s = stack
	return stack, item
}

type StackItem struct {
	i int
	t interface{}
}

type Stack struct {
	stack []StackItem
}

func (s *Stack) Pop() (int, interface{}) {
	item, stack := s.stack[len(s.stack)-1], s.stack[:len(s.stack)-1]
	s.stack = stack
	return item.i, item.t
}

func GenerateTree(ps *PrimitiveSet, min int, max int, condition GenCondition) []interface{} {
	var expr []interface{}
	height := rand.Intn(max-min) + min

	stack := Stack{
		stack: []StackItem{
			{i: 0, t: ps.RetType},
		},
	}
	for len(stack.stack) != 0 {
		depth, type_ := stack.Pop()
		realType := type_.(reflect.Kind)
		fmt.Printf("Condition at %d is %t\n", len(stack.stack)+1, condition(height, depth, min, max, ps))
		if condition(height, depth, min, max, ps) {
			term := ps.Terminals[realType][rand.Intn(len(ps.Terminals[realType]))]
			if term == nil {
				panic("No terminal with type available")
			}
			//if type(term) is MetaEphemeral:
			//    term = term()
			expr = append(expr, term)
		} else {
			prim := ps.Primitives[realType][rand.Intn(len(ps.Primitives[realType]))]
			if prim == nil {
				panic("No primitive with type available")
			}
			expr = append(expr, prim)
			for i := len(prim.argTypes) - 1; i >= 0; i-- {
				arg := prim.argTypes[i]
				stack.stack = append(stack.stack, StackItem{i: depth + 1, t: arg})
			}
		}
	}
	return expr
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
		ret := 0
		for _, n := range a {
			ret -= n.(int)
		}
		return ret
	}
	p := NewPrimitive(fone, []reflect.Kind{reflect.Int, reflect.Int}, reflect.Int)
	p2 := NewPrimitive(ftwo, []reflect.Kind{reflect.Int, reflect.Int}, reflect.Int)
	t := NewTerminal(reflect.Int, 99)
	// fmt.Println(p.Eval(8, 7))
	// fmt.Println(p2.Eval(3, 4))

	ps := NewPrimitiveSet([]reflect.Kind{reflect.Int, reflect.Int}, reflect.Int)
	ps.AddPrimitive(p)
	ps.AddPrimitive(p2)
	ps.AddTerminal(t)
	fmt.Printf("Terminals: %+v Primitives: %+v\n", ps.Terminals, ps.Primitives)
	ret := GenerateTree(ps, 2, 3, GenFull)
	fmt.Printf("tree: %+v\n", ret[1])
}

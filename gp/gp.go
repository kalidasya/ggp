package gp

import (
	"errors"
	"fmt"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"math/rand"
	"reflect"
	"strings"
	"time"
)

type PrimitiveFunc func(...interface{}) interface{}

type GenCondition func(int, int, int, int, *PrimitiveSet) bool

var GenGrow GenCondition = func(height int, depth int, min int, max int, ps *PrimitiveSet) bool {
	return depth == height || (depth >= min && rand.Float32() < ps.TerminalRatio())
}

var GenFull GenCondition = func(height int, depth int, _min int, _max int, _ps *PrimitiveSet) bool {
	return depth == height
}

// ------- Primitive Tree

type NodeString struct {
	node Node
	str  []string
}

type NodeInterace struct {
	node Node
	args []interface{}
}

type PrimitiveTree struct {
	stack []Node // either primitive or terminal
}

// somehow the last node is not a terminal, something is wrong with the tree growing

func (pt *PrimitiveTree) String() string {
	var stack []NodeString
	for _, node := range pt.stack {
		// fmt.Printf("!!pt.stack len: %d stack len: %d c: %d\n", len(pt.stack), len(stack), i)
		stack = append(stack, NodeString{node, []string{}})
		// fmt.Printf("!!Node: %s %d==%d\n", node.Name(), len(stack[len(stack)-1].str), stack[len(stack)-1].node.Arity())
		for len(stack[len(stack)-1].str) == stack[len(stack)-1].node.Arity() {
			var n NodeString
			stack, n = Pop(stack)
			s := n.node.Str(n.str)
			// fmt.Printf("!!current str: %s\n", s)
			if len(stack) == 0 {
				return s
			}
			stack[len(stack)-1].str = append(stack[len(stack)-1].str, s)
		}
	}
	return "."
}

func (pt *PrimitiveTree) Compile() interface{} {
	var stack []NodeInterace
	for _, node := range pt.stack {
		stack = append(stack, NodeInterace{node, []interface{}{}})
		// fmt.Printf("Node: %s %d==%d\n", node.Name(), len(stack[len(stack)-1].args), stack[len(stack)-1].node.Arity())
		for len(stack[len(stack)-1].args) == stack[len(stack)-1].node.Arity() {
			var n NodeInterace
			stack, n = Pop(stack)
			// fmt.Printf("last stack %s %v\n", n.node.Name(), n.args)
			res, err := n.node.Eval(n.args)
			if err != nil {
				panic("eval error")
			}
			if len(stack) == 0 {
				return res
			}
			stack[len(stack)-1].args = append(stack[len(stack)-1].args, res)
		}
	}
	return nil
}

func (pt *PrimitiveTree) Root() interface{} {
	return pt.stack[0]
}

func (pt *PrimitiveTree) Nodes() []Node {
	return pt.stack
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

// ------- Nodes

type Node interface {
	Arity() int
	Name() string
	Eval([]interface{}) (interface{}, error)
	Str([]string) string
	Ret() reflect.Kind
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

func (t *Terminal) Eval(_ []interface{}) (interface{}, error) {
	return t.value, nil
}

func (t *Terminal) Str(_ []string) string {
	return fmt.Sprintf("%v", t.value)
}

func (t *Terminal) Ret() reflect.Kind {
	return t.retType
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

func (p *Primitive) Eval(args []interface{}) (interface{}, error) {
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

func (p *Primitive) Str(args []string) string {
	return fmt.Sprintf("%s(%s)", p.Name(), strings.Join(args, ", "))
}

func (p *Primitive) Ret() reflect.Kind {
	return p.retType
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

// -------------- PrimitiveSet

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

type StackItem struct {
	i int
	t reflect.Kind
}

func GenerateTree(ps *PrimitiveSet, min int, max int, condition GenCondition, type_ reflect.Kind) *PrimitiveTree {
	var expr []Node
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	height := r.Intn(max-min) + min
	fmt.Printf("Generated height: %d\n", height)

	if type_ == reflect.Invalid {
		type_ = ps.RetType
	}

	stack := []StackItem{
		{i: 0, t: type_},
	}

	for len(stack) != 0 {
		var item StackItem
		stack, item = Pop(stack)
		depth := item.i
		realType := item.t
		if condition(height, depth, min, max, ps) {
			term := ps.Terminals[realType][r.Intn(len(ps.Terminals[realType]))]
			if term == nil {
				panic("No terminal with type available")
			}
			expr = append(expr, term)
		} else {
			prim := ps.Primitives[realType][r.Intn(len(ps.Primitives[realType]))]
			if prim == nil {
				panic("No primitive with type available")
			}
			expr = append(expr, prim)
			for i := len(prim.argTypes) - 1; i >= 0; i-- {
				stack = append(stack, StackItem{i: depth + 1, t: prim.argTypes[i]})
			}
		}
	}
	return NewPrimitiveTree(expr)
}

func CXOnePoint(ind1 *PrimitiveTree, ind2 *PrimitiveTree) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if len(ind1.stack) < 2 || len(ind2.stack) < 2 {
		return
	}

	types1 := make(map[reflect.Kind][]int)
	for i, n := range ind1.stack[1:] {
		types1[n.Ret()] = append(types1[n.Ret()], i+1)
	}
	types2 := make(map[reflect.Kind][]int)
	for i, n := range ind2.stack[1:] {
		types2[n.Ret()] = append(types2[n.Ret()], i+1)
	}

	// todo refactor to set creation and intersection
	type1Keys := maps.Keys(types1)
	slices.Sort(type1Keys)
	slices.Compact(type1Keys)

	type2Keys := maps.Keys(types2)
	slices.Sort(type2Keys)
	slices.Compact(type2Keys)

	var commonTypes []reflect.Kind
	for _, t1 := range type1Keys {
		for _, t2 := range type2Keys {
			if t1 == t2 {
				commonTypes = append(commonTypes, t1)
			}
		}
	}
	if len(commonTypes) > 0 {
		type_ := commonTypes[r.Intn(len(commonTypes))]

		index1 := types1[type_][r.Intn(len(types1[type_]))]
		index2 := types2[type_][r.Intn(len(types2[type_]))]

		slice1Begin, slice1End := ind1.SearchSubtree(index1)
		slice2Begin, slice2End := ind2.SearchSubtree(index2)
		ind1.stack = replaceInRange(ind1.stack, slice1Begin, slice1End, ind2.stack[slice2Begin:slice2End]...)
		ind2.stack = replaceInRange(ind2.stack, slice2Begin, slice2End, ind1.stack[slice1Begin:slice1End]...)
	}
}

func MutUniform(ind *PrimitiveTree, expr func(*PrimitiveSet, reflect.Kind) []Node, ps *PrimitiveSet) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	index := r.Intn(len(ind.stack))
	sliceStart, sliceEnd := ind.SearchSubtree(index)
	type_ := ind.stack[index].Ret()
	newNodes := expr(ps, type_)
	fmt.Printf("mutation from %d to %d adding: %d nodes\n", sliceStart, sliceEnd, len(newNodes))
	ind.stack = replaceInRange(ind.stack, sliceStart, sliceEnd, newNodes...)
}

func replaceInRange(stack []Node, start, end int, insert ...Node) []Node {
	stack = slices.Delete(stack, start, end)
	stack = slices.Insert(stack, start, insert...)
	return stack
}

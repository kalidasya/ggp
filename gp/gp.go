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
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return depth == height || (depth >= min && r.Float32() < ps.TerminalRatio())
}

var GenFull GenCondition = func(height int, depth int, _min int, _max int, _ps *PrimitiveSet) bool {
	return depth == height
}

var GenHalfAndHalf GenCondition = func(height int, depth int, min int, max int, ps *PrimitiveSet) bool {
	if rand.Intn(2) == 0 {
		return GenGrow(height, depth, min, max, ps)
	}
	return GenFull(height, depth, min, max, ps)
}

// ------- Primitive Tree

type NodeString struct {
	node Node
	str  []string
}

type NodeInterface struct {
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
		stack = append(stack, NodeString{node, []string{}})
		for len(stack[len(stack)-1].str) == stack[len(stack)-1].node.Arity() {
			var n NodeString
			stack, n = Pop(stack)
			s := n.node.Str(n.str)
			if len(stack) == 0 {
				return s
			}
			stack[len(stack)-1].str = append(stack[len(stack)-1].str, s)
		}
	}
	return "."
}

func (pt *PrimitiveTree) Compile(arguments ...interface{}) interface{} {
	var stack []NodeInterface
  argumentsMap := make(map[string]interface{})
  for i, a := range arguments {
    argumentsMap[fmt.Sprintf("__ARG__%d", i)] = a
  }
	for _, node := range pt.stack {
		stack = append(stack, NodeInterface{node, []interface{}{}})
		for len(stack[len(stack)-1].args) == stack[len(stack)-1].node.Arity() {
			var n NodeInterface
      var res interface{}
      var err error
			stack, n = Pop(stack)
      // here we pass the received values for each argument terminal
      if ind := slices.Index(maps.Keys(argumentsMap), n.node.Name()); ind >-1 {
        // argument terminals are always receiving a single value but the interface requires a list
        res, err = n.node.Eval([]interface{}{argumentsMap[n.node.Name()]})
      } else {
  			res, err = n.node.Eval(n.args)
      }
			if err != nil {
				fmt.Println(err.Error())
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
	fmt.Printf("begin node %s arity is %d in index %d\n", pt.stack[begin].Name(), total, begin)
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
	String() string
}

type Terminal struct {
	name    string
	retType reflect.Kind
	value   interface{}
  argument bool
}

func (t *Terminal) Arity() int {
	return 0
}

func (t *Terminal) Name() string {
	return t.name
}

func (t *Terminal) Eval(argValues []interface{}) (interface{}, error) {
  if t.argument {
    if len(argValues) != 1 {
      return nil, errors.New("argument terminal can only have one return value")
    }
    return argValues[0], nil
  }
	return t.value, nil
}

func (t *Terminal) Str(_ []string) string {
  if t.argument {
    return t.Name()
  }
	switch t.retType {
	case reflect.String:
		return fmt.Sprintf(`"%s"`, t.value)
	default:
		return fmt.Sprintf("%v", t.value)
	}
}

func (t *Terminal) Ret() reflect.Kind {
	return t.retType
}

func (t *Terminal) String() string {
	return t.name
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
			return nil, errors.New(fmt.Sprintf("invalid type for %dth argument", i+1))
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

func (p *Primitive) String() string {
	return p.name
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
	prims := ps.Primitives[p.retType]
	ps.Primitives[p.Ret()] = append(prims, p)
}

func (ps *PrimitiveSet) AddTerminal(t *Terminal) {
	terms := ps.Terminals[t.retType]
	ps.Terminals[t.retType] = append(terms, t)
}

func (ps *PrimitiveSet) TerminalRatio() float32 {
	return float32(len(ps.Terminals)) / float32(len(ps.Terminals)+len(ps.Primitives))
}

// TODO input types are ignored, no symbolic terminal
func NewPrimitiveSet(inTypes []reflect.Kind, retType reflect.Kind) *PrimitiveSet {
  ps := &PrimitiveSet{
		Primitives: make(map[reflect.Kind][]*Primitive),
		Terminals:  make(map[reflect.Kind][]*Terminal),
		RetType:    retType,
		InTypes:    inTypes,
		Arity:      len(inTypes),
	}
  
  for i, r := range inTypes {
    argName := fmt.Sprintf("__ARG__%d", i)
    inTerminal := &Terminal{
      name: argName,
      retType: r,
      argument: true,
    }
    ps.AddTerminal(inTerminal)
  }
	return ps
}

type StackItem struct {
	i int
	t reflect.Kind
}

func GenerateTree(ps *PrimitiveSet, min int, max int, condition GenCondition, type_ reflect.Kind, r *rand.Rand) *PrimitiveTree {
	var expr []Node
	height := r.Intn(max-min) + min

	stack := []StackItem{
		{i: 0, t: type_},
	}

	for len(stack) != 0 {
		var item StackItem
		stack, item = Pop(stack)
		depth, realType := item.i, item.t
		// realType := item.t
		if condition(height, depth, min, max, ps) {
			term := ps.Terminals[realType][r.Intn(len(ps.Terminals[realType]))]
			if term == nil {
				panic("No terminal with type available") // assert.Panics
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

type CrossOver func(*PrimitiveTree, *PrimitiveTree, *rand.Rand, int) (*PrimitiveTree, *PrimitiveTree)
// type CrossOverLimiter func(CrossOver, any) CrossOver

func StaticCrossOverLimiter(crossover CrossOver, limit int) CrossOver {
  return func(ind1 *PrimitiveTree, ind2 *PrimitiveTree, r *rand.Rand, bias int) (*PrimitiveTree, *PrimitiveTree) {
    child1, child2 := crossover(ind1, ind2, r, bias)
    parents := []*PrimitiveTree{ind1, ind2}
    if len(child1.Nodes()) > limit {
      child1 = parents[rand.Intn(len(parents))]
    }
    if len(child2.Nodes()) > limit {
      child2 = parents[rand.Intn(len(parents))]
    }
    return child1, child2
  }
}

func CXOnePoint(ind1 *PrimitiveTree, ind2 *PrimitiveTree, r *rand.Rand, _ int) (*PrimitiveTree, *PrimitiveTree) {
	if len(ind1.stack) < 2 || len(ind2.stack) < 2 {
		return nil, nil
	}

	types1 := make(map[reflect.Kind][]int)
	for i, n := range ind1.stack[1:] {
		types1[n.Ret()] = append(types1[n.Ret()], i+1)
	}
	types2 := make(map[reflect.Kind][]int)
	for i, n := range ind2.stack[1:] {
		types2[n.Ret()] = append(types2[n.Ret()], i+1)
	}

	commonTypes := Intersect(maps.Keys(types1), maps.Keys(types2))

	if len(commonTypes) > 0 {
		type_ := commonTypes[r.Intn(len(commonTypes))]
		index1 := types1[type_][r.Intn(len(types1[type_]))]
		index2 := types2[type_][r.Intn(len(types2[type_]))]

		slice1Begin, slice1End := ind1.SearchSubtree(index1)
		slice2Begin, slice2End := ind2.SearchSubtree(index2)

		child1Stack := ReplaceInRange(ind1.stack, slice1Begin, slice1End, ind2.stack[slice2Begin:slice2End]...)
		child2Stack := ReplaceInRange(ind2.stack, slice2Begin, slice2End, ind1.stack[slice1Begin:slice1End]...)
		return NewPrimitiveTree(child1Stack), NewPrimitiveTree(child2Stack)
	}
  return nil, nil
}

// type Mutator func (*PrimitiveTree) *PrimitiveTree
type Mutator func (*PrimitiveTree) *PrimitiveTree

type MutatorLimiter func(Mutator) Mutator

func StaticMutatorLimiter(mutator Mutator, limit int) Mutator {
  return func(ind *PrimitiveTree) *PrimitiveTree {
    res := mutator(ind)
    if len(res.Nodes()) > limit {
      res = ind
    }
    return res
  } 
}

//todo mutNodeReplacement mutEphemeral mutInsert mutShrink

type UniformMutator struct {
  expr func(*PrimitiveSet, reflect.Kind) []Node
  ps *PrimitiveSet
  r *rand.Rand
}

func NewUniformMutator(ps *PrimitiveSet, expr func(*PrimitiveSet, reflect.Kind) []Node, r *rand.Rand) *UniformMutator {
  return &UniformMutator {
    expr: expr,
    r: r,
    ps: ps,
  }
}

func (m *UniformMutator) Mutate(ind *PrimitiveTree) *PrimitiveTree {
	index := m.r.Intn(len(ind.stack))
	sliceStart, sliceEnd := ind.SearchSubtree(index)
	type_ := ind.stack[index].Ret()
	newNodes := m.expr(m.ps, type_)
	return NewPrimitiveTree(ReplaceInRange(ind.stack, sliceStart, sliceEnd, newNodes...))
}

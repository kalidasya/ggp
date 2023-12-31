package ant

import (
	"fmt"
	"io/ioutil"
	"main/gp"
	"math/rand"
	"reflect"
	"strings"
	"time"

	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
)

type direction int
type fieldValue int

const (
	North direction = iota
	East
	South
	West
)

const (
	Empty fieldValue = iota
	Food
	Passed
)

type Direction struct {
	Row int
	Col int
}

func PyMod[T constraints.Integer](d, m T) T {
	if d < 0 && m < 0 {
		return d % m
	}
	d %= m
	if d < 0 {
		d += m
	}
	return d
}

var (
	Directions = map[direction]Direction{
		North: {
			Row: 1,
			Col: 0,
		},
		East: {
			Row: 0,
			Col: 1,
		},
		South: {
			Row: -1,
			Col: 0,
		},
		West: {
			Row: 0,
			Col: -1,
		},
	}
)

type AntIndividual struct {
	tree    *gp.PrimitiveTree
	fitness *gp.Fitness
}

func (a *AntIndividual) Fitness() *gp.Fitness {
	return a.fitness
}

func (a *AntIndividual) Tree() *gp.PrimitiveTree {
	return a.tree
}

func (a *AntIndividual) Copy() gp.Individual {
	fit, err := gp.NewFitness(a.Fitness().GetWeights())
	fit.SetValues(a.Fitness().GetValues())
	if err != nil {
		panic(err)
	}
	return &AntIndividual{
		tree:    gp.NewPrimitiveTree(a.Tree().Nodes()),
		fitness: fit,
	}
}

var _ gp.Individual = new(AntIndividual)

type Matrix struct {
	Rows     int
	Cols     int
	data     [][]fieldValue
	StartRow int
	StartCol int
}

func (m *Matrix) Get(r int, c int) fieldValue {
	return m.data[r][c]
}

func (m *Matrix) Set(r int, c int, v fieldValue) {
	m.data[r][c] = v
}

func (m *Matrix) String() string {
	var ret strings.Builder
	for r := range m.data {
		for c := range m.data[r] {
			switch m.data[r][c] {
			case Empty:
				ret.WriteRune('.')
			case Food:
				ret.WriteRune('#')
			case Passed:
				ret.WriteRune('X')
			}
		}
		ret.WriteRune('\n')
	}
	return ret.String()
}

func (m *Matrix) Copy() Matrix {
	newData := make([][]fieldValue, len(m.data))
	copy(newData, m.data)
	for i := range m.data {
		newData[i] = make([]fieldValue, len(m.data[i]))
		copy(newData[i], m.data[i])
	}
	return Matrix{
		Rows:     m.Rows,
		Cols:     m.Cols,
		StartRow: m.StartRow,
		StartCol: m.StartCol,
		data:     newData,
	}

}

type Ant struct {
	maxMoves       int
	moves          int
	eaten          int
	row            int
	col            int
	dir            direction
	matrix         Matrix
	originalMatrix Matrix
}

func NewAnt(maxMoves int, matrix Matrix) *Ant {
	return &Ant{
		maxMoves:       maxMoves,
		moves:          0,
		eaten:          0,
		dir:            1,
		matrix:         matrix,
		row:            matrix.StartRow,
		col:            matrix.StartCol,
		originalMatrix: matrix.Copy(),
	}
}

func (a *Ant) Reset() {
	a.row = a.matrix.StartRow
	a.col = a.matrix.StartCol
	a.dir = East
	a.moves = 0
	a.eaten = 0
	a.matrix = a.originalMatrix.Copy()
}

func (a *Ant) Position() (int, int, direction) {
	return a.row, a.col, a.dir
}

func (a *Ant) senseFood() bool {
	aheadRow := PyMod(a.row+Directions[a.dir].Row, a.matrix.Rows)
	aheadCol := PyMod(a.col+Directions[a.dir].Col, a.matrix.Cols)
	return a.matrix.Get(aheadRow, aheadCol) == Food
}

func (a *Ant) IfFoodAhead(args ...gp.PrimitiveArgs) gp.PrimitiveArgs {
	return func(subArgs ...gp.PrimitiveArgs) gp.PrimitiveArgs {
		if a.senseFood() {
			return call(args[0], subArgs...)
		}
		return call(args[1], subArgs...)
	}
}

func (a *Ant) TurnLeft(_ ...gp.PrimitiveArgs) gp.PrimitiveArgs {
	if a.moves < a.maxMoves {
		a.moves++
		a.dir = PyMod(a.dir-1, 4)
	}
	return nil
}

func (a *Ant) TurnRight(_ ...gp.PrimitiveArgs) gp.PrimitiveArgs {
	if a.moves < a.maxMoves {
		a.moves++
		a.dir = PyMod(a.dir+1, 4)
	}
	return nil
}

func (a *Ant) MoveForward(_ ...gp.PrimitiveArgs) gp.PrimitiveArgs {
	if a.moves < a.maxMoves {
		a.moves++
		a.row = PyMod(a.row+Directions[a.dir].Row, a.matrix.Rows)
		a.col = PyMod(a.col+Directions[a.dir].Col, a.matrix.Cols)

		if a.matrix.Get(a.row, a.col) == Food {
			a.eaten++
		}
		a.matrix.Set(a.row, a.col, Passed)
	}
	return nil
}

func ProgN(progs ...gp.PrimitiveArgs) gp.PrimitiveArgs {
	return func(args ...gp.PrimitiveArgs) gp.PrimitiveArgs {
		for _, p := range progs {
			call(p, args...)
		}
		return nil
	}

}

type FPP = func(...gp.PrimitiveArgs) gp.PrimitiveArgs

func call(f gp.PrimitiveArgs, args ...gp.PrimitiveArgs) gp.PrimitiveArgs {
	return f.(FPP)(args...)
}

func ParseMatrix(input string) (Matrix, error) {
	bytesRead, err := ioutil.ReadFile(input)
	if err != nil {
		return Matrix{}, err
	}
	ret := Matrix{}
	fileContent := string(bytesRead)
	ret.Rows = strings.Count(fileContent, "\n") + 1
	ret.data = make([][]fieldValue, ret.Rows)
	for r, line := range strings.Split(fileContent, "\n") {
		ret.Cols = len(line)
		ret.data[r] = make([]fieldValue, ret.Cols)
		for c, char := range line {
			switch char {
			case '#':
				ret.data[r][c] = Food
				continue
			case 'S':
				ret.StartRow = r
				ret.StartCol = c
				ret.data[r][c] = Empty
			default:
				ret.data[r][c] = Empty
			}
		}
	}
	return ret, nil
}

func eval(ant *Ant, ind gp.Individual) {
	ant.Reset()
	routine := ind.Tree().Compile().(func(...gp.PrimitiveArgs) gp.PrimitiveArgs)
	for ant.moves < ant.maxMoves {
		routine()
	}
	ind.Fitness().SetValues([]float32{float32(ant.eaten)})
}

func Main() {
	/*
	  Best in gen: [84.00]
	  best algo:
	  prog3(prog3(turn_right, move_forward, turn_left), prog3(if_food_ahead(move_forward, if_food_ahead(prog3(turn_left, if_food_ahead(move_forward, move_forward), turn_right), turn_left)), if_food_ahead(turn_left, move_forward), if_food_ahead(if_food_ahead(prog2(move_forward, move_forward), turn_left), turn_right)), move_forward)
	*/
	matrix, err := ParseMatrix("examples/ant/matrix.txt")
	if err != nil {
		panic(err)
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	ant := NewAnt(600, matrix)

	inds := []gp.Individual{}

	ps := gp.NewPrimitiveSet([]reflect.Kind{}, reflect.Func)
	ps.AddPrimitive(gp.NewPrimitive("prog3", ProgN, []reflect.Kind{reflect.Func, reflect.Func, reflect.Func}, reflect.Func))
	ps.AddPrimitive(gp.NewPrimitive("prog2", ProgN, []reflect.Kind{reflect.Func, reflect.Func}, reflect.Func))
	ps.AddPrimitive(gp.NewPrimitive("if_food_ahead", ant.IfFoodAhead, []reflect.Kind{reflect.Func, reflect.Func}, reflect.Func))
	ps.AddTerminal(gp.NewTerminal("move_forward", reflect.Func, ant.MoveForward))
	ps.AddTerminal(gp.NewTerminal("turn_left", reflect.Func, ant.TurnLeft))
	ps.AddTerminal(gp.NewTerminal("turn_right", reflect.Func, ant.TurnRight))

	for i := 0; i < 300; i++ {
		fit, err := gp.NewFitness([]float32{1})
		if err != nil {
			panic(err)
		}
		ind := AntIndividual{
			tree:    gp.GenerateTree(ps, 1, 2, gp.GenFull, ps.RetType, r),
			fitness: fit,
		}
		inds = append(inds, &ind)
	}

	settings := gp.AlgorithmSettings{
		NumGen:               40,
		MutationProbability:  0.5,
		CrossoverProbability: 0.2,
		SelectionSize:        len(inds),
		TournamentSize:       7,
		CrossOverFunc:        gp.CXOnePoint,
		MutatorFunc: gp.NewUniformMutator(ps, func(ps *gp.PrimitiveSet, type_ reflect.Kind) []gp.Node {
			return gp.GenerateTree(ps, 0, 2, gp.GenFull, type_, r).Nodes()
		}, r).Mutate,
	}
	inds = gp.EaSimple(inds, ps, func(ind gp.Individual) {
		eval(ant, ind)
	}, settings, r)
	best := slices.MaxFunc(inds, gp.FitnessMaxFunc)
	eval(ant, best)
	fmt.Printf("best algo: \n%s\n", best.Tree().String())
	fmt.Printf("best matrix: \n%s\n", ant.matrix.String())

}

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

var _ gp.Individual = new(AntIndividual)

type Matrix struct {
	Rows     int
	Cols     int
	data     [][]fieldValue
	StartRow int
	StartCol int
}

func (m *Matrix) Get(r int, c int) fieldValue {
	// fmt.Printf("Getting %d:%d from matrix %dx%d\n", r, c, m.Rows, m.Cols)
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

// type Move func() interface{} // we need return values

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
		originalMatrix: matrix,
	}
}

func (a *Ant) Reset() {
	a.row = a.matrix.StartRow
	a.col = a.matrix.StartCol
	a.dir = East
	a.moves = 0
	a.eaten = 0
	a.matrix = a.originalMatrix
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
	if a.senseFood() {
		return func(subArgs ...gp.PrimitiveArgs) gp.PrimitiveArgs {
			return call(args[0], subArgs...)
		}
	} else {
		return func(subArgs ...gp.PrimitiveArgs) gp.PrimitiveArgs {
			return call(args[1], subArgs...)
		}
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

// TODO this is shite
type FPFP func(...gp.PrimitiveArgs) func(...gp.PrimitiveArgs)
type FPP func(...gp.PrimitiveArgs) gp.PrimitiveArgs
type FP func(...gp.PrimitiveArgs)
type F func()

func call(f gp.PrimitiveArgs, args ...gp.PrimitiveArgs) gp.PrimitiveArgs {
	fTypeCheck := reflect.TypeOf(f).ConvertibleTo
	if fTypeCheck(reflect.TypeOf((FPFP)(nil))) {
		return call(f.(func(...gp.PrimitiveArgs) func(...gp.PrimitiveArgs))(args...), args...)
	} else if fTypeCheck(reflect.TypeOf((FPP)(nil))) {
		return f.(func(...gp.PrimitiveArgs) gp.PrimitiveArgs)(args...)
	} else if fTypeCheck(reflect.TypeOf((FP)(nil))) {
		f.(func(...gp.PrimitiveArgs))(args...)
	} else if fTypeCheck(reflect.TypeOf((F)(nil))) {
		f.(func())()
	} else {
		fmt.Printf("Non matching signature: %+v\n", reflect.TypeOf(f))
	}
	return nil
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
	// repeate it until it runs out of moves
	for ant.moves < ant.maxMoves {
		routine()
	}
	ind.Fitness().SetValues([]float32{float32(ant.eaten)})
}

func Main() {
	matrix, err := ParseMatrix("examples/ant/matrix.txt")
	if err != nil {
		panic(err)
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	ant := NewAnt(600, matrix)

	inds := []gp.Individual{}

	ps := gp.NewPrimitiveSet([]reflect.Kind{}, reflect.Func)
	ps.AddPrimitive(gp.NewPrimitive("if_food_ahead", ant.IfFoodAhead, []reflect.Kind{reflect.Func, reflect.Func}, reflect.Func))
	ps.AddPrimitive(gp.NewPrimitive("prog2", ProgN, []reflect.Kind{reflect.Func, reflect.Func}, reflect.Func))
	ps.AddPrimitive(gp.NewPrimitive("prog3", ProgN, []reflect.Kind{reflect.Func, reflect.Func, reflect.Func}, reflect.Func))
	ps.AddTerminal(gp.NewTerminal("move_forward", reflect.Func, ant.MoveForward))
	ps.AddTerminal(gp.NewTerminal("turn_left", reflect.Func, ant.TurnLeft))
	ps.AddTerminal(gp.NewTerminal("turn_right", reflect.Func, ant.TurnRight))

	for i := 0; i < 10; i++ {
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
	gp.EaSimple(inds, ps, func(ind gp.Individual) {
		eval(ant, ind)
	}, r)
}

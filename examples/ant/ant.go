package ant

import (
	"fmt"
	"io/ioutil"
	"main/gp"
	"math/rand"
	"reflect"
	"strings"
	"time"
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

type Individual struct {
	gp.PrimitiveTree
	fitness gp.Fitness
}

type Matrix struct {
	Rows     int
	Cols     int
	data     [][]fieldValue
	StartRow int
	StartCol int
}

func (m *Matrix) Get(r int, c int) fieldValue {
	fmt.Printf("Getting %d:%d from matrix %dx%d\n", r, c, m.Rows, m.Cols)
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
				ret.WriteRune(' ')
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
	routine        Individual
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

func (a *Ant) TurnLeft() interface{} {
	if a.moves < a.maxMoves {
		a.moves++
		a.dir = ((a.dir - 1) % 4) + 4
	}
	return nil
}

func (a *Ant) TurnRight() interface{} {
	if a.moves < a.maxMoves {
		a.moves++
		a.dir = ((a.dir + 1) % 4) + 4
	}
	return nil
}

func (a *Ant) MoveForward() interface{} {
	if a.moves < a.maxMoves {
		a.moves++
		a.row = ((a.row + Directions[a.dir].Row) % a.matrix.Rows) + a.matrix.Rows
		a.col = ((a.col + Directions[a.dir].Col) % a.matrix.Cols) + a.matrix.Cols

		fmt.Printf("current row %d current col %d row direction %d col direction %d matrix rows %d matrix cols %d\n", a.row, a.col,
			Directions[a.dir].Row, Directions[a.dir].Col, a.matrix.Rows, a.matrix.Cols)

		if a.matrix.Get(a.row, a.col) == Food {
			a.eaten++
		}
		a.matrix.Set(a.row, a.col, Passed)
	}
	var ret interface{}
	ret = 1
	return ret
}

func (a *Ant) senseFood() bool {
	fmt.Printf("current row %d current col %d row direction %d col direction %d matrix rows %d matrix cols %d\n", a.row, a.col,
		Directions[a.dir].Row, Directions[a.dir].Col, a.matrix.Rows, a.matrix.Cols)
	aheadRow := ((a.row + Directions[a.dir].Row) % a.matrix.Rows) + a.matrix.Rows
	aheadCol := ((a.col + Directions[a.dir].Col) % a.matrix.Cols) + a.matrix.Cols
	return a.matrix.Get(aheadRow, aheadCol) == Food
}

func (a *Ant) IfFoodAhead(outs ...interface{}) interface{} {
	if a.senseFood() {
		return outs[0].(func() interface{})()
	}
	return outs[1].(func() interface{})()
}

func ProgN(progs ...interface{}) interface{} {
	for _, p := range progs {
		p.(func() interface{})()
	}
	var ret interface{}
	ret = 1
	return ret
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

func Main() {
	matrix, err := ParseMatrix("examples/ant/matrix.txt")
	if err != nil {
		panic(err)
	}
	ant := NewAnt(10, matrix)
	ps := gp.NewPrimitiveSet([]reflect.Kind{}, reflect.Interface)
	ps.AddPrimitive(gp.NewPrimitive("if_food_ahead", ant.IfFoodAhead, []reflect.Kind{reflect.Interface, reflect.Interface}, reflect.Interface))
	ps.AddPrimitive(gp.NewPrimitive("prog2", ProgN, []reflect.Kind{reflect.Interface, reflect.Interface}, reflect.Interface))
	ps.AddPrimitive(gp.NewPrimitive("prog3", ProgN, []reflect.Kind{reflect.Interface, reflect.Interface, reflect.Interface}, reflect.Interface))
	ps.AddTerminal(gp.NewTerminal("move_forward", reflect.Interface, func() interface{} {
		ant.IfFoodAhead()
		var ret interface{}
		ret = 1
		return ret
	}))
	ps.AddTerminal(gp.NewTerminal("turn_left", reflect.Interface, func() interface{} {
		ant.TurnLeft()
		var ret interface{}
		ret = 1
		return ret
	}))
	ps.AddTerminal(gp.NewTerminal("turn_right", reflect.Interface, func() interface{} {
		ant.TurnRight()
		var ret interface{}
		ret = 1
		return ret
	}))

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	ret := gp.GenerateTree(ps, 3, 4, gp.GenFull, ps.RetType, r)
	fmt.Printf("%s\n", ret)
	fmt.Printf("%s\n", ant.matrix.String())
	fmt.Printf("ANT before: %+v\n", ant)
	ret.Compile()
	fmt.Printf("ANT after: %+v\n", ant)

}

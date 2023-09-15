package main

import (
	"fmt"
	"main/gp"
	"math/rand"
	"reflect"
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
	Rows int
	Cols int
	data [][]fieldValue
}

func (m *Matrix) Get(r int, c int) fieldValue {
	return m.data[r][c]
}

func (m *Matrix) Set(r int, c int, v fieldValue) {
	m.data[r][c] = v
}

type Move func() interface{} // we need return values

type Ant struct {
	maxMoves int
	moves    int
	eaten    int
	routine  Individual
	row      int
	col      int
	dir      direction
	rowStart int
	colStart int
	matrix   Matrix
}

func NewAnt(maxMoves int) *Ant {
	return &Ant{
		maxMoves: maxMoves,
		moves:    0,
		eaten:    0,
	}
}

func (a *Ant) Reset() {
	a.row = a.rowStart
	a.col = a.colStart
	a.dir = East
	a.moves = 0
	a.eaten = 0
	// matrix
}

func (a *Ant) Position() (int, int, direction) {
	return a.row, a.col, a.dir
}

func (a *Ant) TurnLeft() interface{} {
	if a.moves < a.maxMoves {
		a.moves++
		a.dir = (a.dir - 1) % 4
	}
	return nil
}

func (a *Ant) TurnRight() interface{} {
	if a.moves < a.maxMoves {
		a.moves++
		a.dir = (a.dir + 1) % 4
	}
	return nil
}

func (a *Ant) MoveForward() interface{} {
	if a.moves < a.maxMoves {
		a.moves++
		a.row = (a.row + Directions[a.dir].Row) % a.matrix.Rows
		a.col = (a.col + Directions[a.dir].Col) % a.matrix.Cols
		if a.matrix.Get(a.row, a.col) == Food {
			a.eaten++
		}
		a.matrix.Set(a.row, a.col, Passed)
	}
	return nil
}

func (a *Ant) senseFood() bool {
	aheadRow := (a.row + Directions[a.dir].Row) % a.matrix.Rows
	aheadCol := (a.col + Directions[a.dir].Col) % a.matrix.Cols
	return a.matrix.Get(aheadRow, aheadCol) == Food
}

func (a *Ant) IfFoodAhead(outs ...interface{}) interface{} {
	if a.senseFood() {
		return outs[0].(Move)()
	}
	return outs[1].(Move)()
}

func ProgN(progs ...interface{}) interface{} {
	for _, p := range progs {
		p.(Move)()
	}
	return nil
}

func antMain() {
	ant := NewAnt(10)
	ps := gp.NewPrimitiveSet([]reflect.Kind{}, reflect.Interface)
	ps.AddPrimitive(gp.NewPrimitive("if_food_ahead", ant.IfFoodAhead, []reflect.Kind{reflect.Interface, reflect.Interface}, reflect.Interface))
	ps.AddPrimitive(gp.NewPrimitive("prog2", ProgN, []reflect.Kind{reflect.Interface, reflect.Interface}, reflect.Interface))
	ps.AddPrimitive(gp.NewPrimitive("prog3", ProgN, []reflect.Kind{reflect.Interface, reflect.Interface, reflect.Interface}, reflect.Interface))
	ps.AddTerminal(gp.NewTerminal("move_forward", reflect.Interface, ant.MoveForward))
	ps.AddTerminal(gp.NewTerminal("turn_left", reflect.Interface, ant.TurnLeft))
	ps.AddTerminal(gp.NewTerminal("turn_right", reflect.Interface, ant.TurnRight))

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	ret := gp.GenerateTree(ps, 3, 4, gp.GenFull, ps.RetType, r)
	fmt.Printf("%s\n", ret)
	fmt.Printf("ANT before: %+v", ant)
	ret.Compile()
	fmt.Printf("ANT after: %+v", ant)

}

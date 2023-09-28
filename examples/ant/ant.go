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

type Individual struct {
	tree    gp.PrimitiveTree
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
	// fmt.Printf("current row %d current col %d row direction %d col direction %d matrix rows %d matrix cols %d\n", a.row, a.col,
	// Directions[a.dir].Row, Directions[a.dir].Col, a.matrix.Rows, a.matrix.Cols)
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
			// fmt.Printf("calling iffood ahead %v no food\n", args[1])
			return call(args[1], subArgs...)
		}
	}
}

func (a *Ant) TurnLeft(_ ...gp.PrimitiveArgs) gp.PrimitiveArgs {
	// return func() {
	// fmt.Println("TurnLeft")
	if a.moves < a.maxMoves {
		a.moves++
		a.dir = PyMod(a.dir-1, 4)
	}
	// }
	return nil
}

func (a *Ant) TurnRight(_ ...gp.PrimitiveArgs) gp.PrimitiveArgs {
	// return func() {
	// fmt.Println("TurnRight")

	if a.moves < a.maxMoves {
		a.moves++
		a.dir = PyMod(a.dir+1, 4)
	}
	// }
	return nil
}

func (a *Ant) MoveForward(_ ...gp.PrimitiveArgs) gp.PrimitiveArgs {
	// return func() {
	// fmt.Println("Moveforward")
	if a.moves < a.maxMoves {
		a.moves++
		// fmt.Printf("start row %d current col %d row direction %d col direction %d matrix rows %d matrix cols %d\n", a.row, a.col,
		// 	Directions[a.dir].Row, Directions[a.dir].Col, a.matrix.Rows, a.matrix.Cols)
		a.row = PyMod(a.row+Directions[a.dir].Row, a.matrix.Rows)
		a.col = PyMod(a.col+Directions[a.dir].Col, a.matrix.Cols)

		// fmt.Printf("end row %d current col %d\n", a.row, a.col)

		if a.matrix.Get(a.row, a.col) == Food {
			a.eaten++
			// fmt.Printf("ant %s eating %d\n", a.id, a.eaten)
		}
		a.matrix.Set(a.row, a.col, Passed)
		// fmt.Printf("Setting %d:%d as passed", a.row, a.col)
	}
	// }
	return nil
}

func ProgN(progs ...gp.PrimitiveArgs) gp.PrimitiveArgs {
	return func(args ...gp.PrimitiveArgs) gp.PrimitiveArgs {
		// fmt.Printf("calling progn %v\n", progs)
		for _, p := range progs {
			// recursive call?
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

func eval(ant *Ant, ind *Individual) {
	ant.Reset()
	// fmt.Printf("Evaling %s\n", ind.tree)
	routine := ind.tree.Compile().(func(...gp.PrimitiveArgs) gp.PrimitiveArgs)
	// repeate it until it runs out of moves
	for ant.moves < ant.maxMoves {
		routine()
	}
	// fmt.Printf("fitness: %d\n", &ind.fitness)
	ind.fitness.SetValues([]float32{float32(ant.eaten)})
	// if ant.eaten > 0 {
	// 	fmt.Printf("ant eval eaten: %d\n", ant.eaten)
	// }
	// fmt.Printf("end evaluating ant %d eaten %d\n", len(ind.tree.Nodes()), ind.ant.eaten)
}

func selRandom(individuals []Individual, k int, r *rand.Rand) []Individual {
	perm := r.Perm(len(individuals))
	chosen := make([]Individual, k)
	for i, randIndex := range perm {
		if i >= k {
			break
		}
		chosen[i] = individuals[randIndex]
	}
	return chosen
}

func selTournament(individuals []Individual, k int, tournsize int, r *rand.Rand) []Individual {
	if k > len(individuals) {
		k = len(individuals)
	}
	chosen := make([]Individual, k)

	for i := 0; i < k; i++ {
		aspirants := selRandom(individuals, tournsize, r)
		sel := slices.MaxFunc(aspirants, func(a, b Individual) int {
			if a.fitness.LessThan(b.fitness) {
				return -1
			} else if a.fitness.Equals(b.fitness) {
				return 0
			}
			return 1
		})
		chosen[i] = sel
		// if val, _ := chosen[i].fitness.GetValues(); val[0] > 0.0 {
		// 	fmt.Printf("Selected %d: %s\n", i, chosen[i].fitness)
		// }
	}
	// fmt.Println("selTournament")
	// for _, o := range chosen {
	// 	fmt.Printf(" %s ", o.fitness)
	// }
	// fmt.Println()
	return chosen
}

func Main() {
	matrix, err := ParseMatrix("examples/ant/matrix.txt")
	if err != nil {
		panic(err)
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	ant := NewAnt(600, matrix)

	inds := []Individual{}

	// ant is a single instance, PS and tree are independent from it and they are evaluated one by one on the ant
	// instance and handled separately. it means we can use the struct functions directly but need to decouple ant from
	// fitness and tree (there can be a single ps as well)
	//

	ps := gp.NewPrimitiveSet([]reflect.Kind{}, reflect.Func)
	ps.AddPrimitive(gp.NewPrimitive("if_food_ahead", ant.IfFoodAhead, []reflect.Kind{reflect.Func, reflect.Func}, reflect.Func))
	ps.AddPrimitive(gp.NewPrimitive("prog2", ProgN, []reflect.Kind{reflect.Func, reflect.Func}, reflect.Func))
	ps.AddPrimitive(gp.NewPrimitive("prog3", ProgN, []reflect.Kind{reflect.Func, reflect.Func, reflect.Func}, reflect.Func))
	ps.AddTerminal(gp.NewTerminal("move_forward", reflect.Func, ant.MoveForward))
	ps.AddTerminal(gp.NewTerminal("turn_left", reflect.Func, ant.TurnLeft))
	ps.AddTerminal(gp.NewTerminal("turn_right", reflect.Func, ant.TurnRight))

	for i := 0; i < 300; i++ {
		fit, err := gp.NewFiness([]float32{1})
		if err != nil {
			panic(err)
		}
		ind := Individual{
			tree:    *gp.GenerateTree(ps, 2, 3, gp.GenFull, ps.RetType, r),
			fitness: *fit,
		}
		// fmt.Printf(" %d %d\n", &ind.fitness, &ind.tree)
		inds = append(inds, ind)
	}

	// fmt.Println("Start")
	// for i := range inds {
	// 	fmt.Printf(" %d %d\n", &inds[i].fitness, &inds[i].tree)
	// }
	// fmt.Println()

	var varAnd = func(inds []Individual, cxpb, mutpb float32) []Individual {
		offs := slices.Clone(inds)
		for i := 1; i < len(offs); i += 2 {
			if rand.Float32() < cxpb {
				// fmt.Printf("crossing ant %s and ant %s\n", offs[i-1].ant.id, offs[i].ant.id)
				offs[i-1].tree, offs[i].tree = gp.CXOnePoint(offs[i-1].tree, offs[i].tree, r, 0)
				// fmt.Printf("CX: %s %s\n", offs[i-1].fitness, offs[i].fitness)
				offs[i-1].fitness.DelValues()
				offs[i].fitness.DelValues()
			}
		}
		// fmt.Println("After CX")
		// for i := range offs {
		// 	fmt.Printf(" %s ", offs[i].fitness)
		// }
		// fmt.Println()
		for i := 0; i < len(offs); i++ {
			uniformMutator := gp.NewUniformMutator(ps, func(ps *gp.PrimitiveSet, type_ reflect.Kind) []gp.Node {
				return gp.GenerateTree(ps, 0, 2, gp.GenFull, type_, r).Nodes()
			}, r)
			if rand.Float32() < mutpb {
				offs[i].tree = *gp.StaticMutatorLimiter(uniformMutator.Mutate, 9999)(&offs[i].tree)
				// fmt.Printf("Mut: %s\n", offs[i].fitness)
				offs[i].fitness.DelValues()
			}
		}
		return offs
	}

	for gen := 0; gen < 80; gen++ {
		fmt.Printf("------------------------------------------------------------------- (%d)\n", gen+1)
		offsprings := selTournament(inds, len(inds), 7, r)
		// fmt.Println("After selTournamen")
		// for i := range offsprings {
		// 	fmt.Printf(" %s ", offsprings[i].fitness.String())
		// }
		// fmt.Println()
		offsprings = varAnd(offsprings, 0.5, 0.2)
		// fmt.Println("After varAnd")
		// for i := range offsprings {
		// 	fmt.Printf(" %s ", offsprings[i].fitness.String())
		// }
		// fmt.Println()
		// fmt.Printf("ANT PATH: \n%s\n", &ant.matrix)
		// fmt.Printf("last algo: \n%s\n", offsprings[len(offsprings)-1].tree)
		// fmt.Printf("last fitness: %s\n", offsprings[len(offsprings)-1].fitness)
		// fmt.Println("Before eval")
		for i := range offsprings {
			// if i > 0 {
			// 	fmt.Printf("fitness i-1 == %t\n", &offsprings[i-1] == &offsprings[i])
			// }
			// if i < len(offsprings)-1 {
			// 	fmt.Printf("fitness %s: %s\n", offsprings[i].fitness.String(), offsprings[i+1].fitness.String())
			// } else {
			// 	fmt.Printf("fitness %s: %s\n", offsprings[i].fitness.String(), offsprings[i].fitness.String())
			// }
			if !offsprings[i].fitness.Valid() {
				// fmt.Print("*")
				eval(ant, &offsprings[i])
			}
			// fitness, err := i.fitness.GetValues()
			// if err != nil {
			// 	panic(err)
			// }
			// if ant.eaten > 0 {
			// 	fmt.Printf("ant eaten %d fitness: %s \n", ant.eaten, i.fitness)
			// 	fmt.Println(i.tree)
			// }
		}
		best := slices.MaxFunc(inds, func(a, b Individual) int {
			if a.fitness.LessThan(b.fitness) {
				return -1
			} else if a.fitness.Equals(b.fitness) {
				return 0
			}
			return 1
		})
		fmt.Printf("Best in gen: %s\n", best.fitness.String())
		// fmt.Println("After eval")
		// for i := range offsprings {
		// 	fmt.Printf(" %s ", offsprings[i].fitness)
		// }
		// fmt.Println()

		// fmt.Println("After re-eval")
		// for _, o := range offsprings {
		// 	fmt.Printf(" %s", o.fitness)
		// }
		// fmt.Println()
		inds = offsprings
	}
	// fmt.Println("Final")
	// for i := range inds {
	// 	fmt.Printf(" %s %s\n", inds[i].fitness, inds[i].tree)
	// }
	// fmt.Println()
	best := slices.MaxFunc(inds, func(a, b Individual) int {
		if a.fitness.LessThan(b.fitness) {
			return -1
		} else if a.fitness.Equals(b.fitness) {
			return 0
		}
		return 1
	})
	fmt.Printf("Func: %s %s", best.fitness.String(), best.tree.String())
	// ret := gp.GenerateTree(ps, 1, 2, gp.GenFull, ps.RetType, r)
	// fmt.Printf("%s\n", ret)
	// fmt.Printf("%s\n", ant.matrix.String())
	// fmt.Printf("ANT before: %+v\n", ant)
	// res := ret.Compile()
	// fmt.Printf("type of res: %+v\n", reflect.TypeOf(res))
	// ret.Compile().(func(...gp.PrimitiveArgs))()
	// fmt.Printf("ANT after: %+v\n", ant)

}

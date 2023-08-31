package main

import (
	"fmt"
	"main/gp"
	"reflect"
)

func main() {
	var fone, ftwo gp.PrimitiveFunc
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
	p := gp.NewPrimitive("fone", fone, []reflect.Kind{reflect.Int, reflect.Int}, reflect.Int)
	p2 := gp.NewPrimitive("ftwo", ftwo, []reflect.Kind{reflect.Int, reflect.Int}, reflect.Int)
	t := gp.NewTerminal("a", reflect.Int, 99)
	t2 := gp.NewTerminal("b", reflect.Int, 12)
	// fmt.Println(p.Eval(8, 7))
	// fmt.Println(p2.Eval(3, 4))

	ps := gp.NewPrimitiveSet([]reflect.Kind{reflect.Int, reflect.Int}, reflect.Int)
	ps.AddPrimitive(p)
	ps.AddPrimitive(p2)
	ps.AddTerminal(t)
	ps.AddTerminal(t2)
	ret := gp.GenerateTree(ps, 3, 4, gp.GenFull)
	fmt.Printf("tree depth: %d\n", ret.Height())
	fmt.Printf("tree: %s\n", ret)
	fmt.Println(ftwo(ftwo(ftwo(12, 99), fone(99, 99)), fone(ftwo(12, 99), fone(99, 99))))
}

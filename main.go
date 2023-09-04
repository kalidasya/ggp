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
	ps := gp.NewPrimitiveSet([]reflect.Kind{reflect.Int, reflect.Int}, reflect.Int)
	ps.AddPrimitive(p)
	ps.AddPrimitive(p2)
	ps.AddTerminal(t)
	ps.AddTerminal(t2)
	r := gp.NewRealRandom()

	ret := gp.GenerateTree(ps, 3, 4, gp.GenFull, ps.RetType, r)
	ret2 := gp.GenerateTree(ps, 3, 4, gp.GenFull, ps.RetType, r)
	// fmt.Printf("tree depth: %d\n", ret.Height())
	// fmt.Printf("new tree: %s=%s\n", ret, ret.Compile())
	// fmt.Printf("new tree2: %s=%s\n", ret2, ret2.Compile())
	// fmt.Println(ftwo(ftwo(ftwo(12, 99), fone(99, 99)), fone(ftwo(12, 99), fone(99, 99))))
	// gp.CXOnePoint(ret, ret2)
	// fmt.Printf("cx tree2: %s=%s\n", ret2, ret2.Compile())
	// fmt.Printf("cx tree: %s=%s\n", ret, ret.Compile())
	count := 100
	for count > 0 && len(fmt.Sprint(ret)) > 1 {
		fmt.Println("===========================================")
		fmt.Printf("Before CX %d:\n", len(ret.Nodes()))
		fmt.Println(ret)
		printNodes(ret)
		fmt.Println("---------------------------------------------")
		fmt.Println(ret2)
		printNodes(ret2)
		gp.CXOnePoint(ret, ret2, r)
		fmt.Printf("After CX before mut %d:\n", len(ret.Nodes()))
		fmt.Println(ret)
		printNodes(ret)
		fmt.Println("---------------------------------------------")
		fmt.Println(ret2)
		printNodes(ret2)

		gp.MutUniform(ret, func(ps *gp.PrimitiveSet, type_ reflect.Kind) []gp.Node {
			return gp.GenerateTree(ps, 0, 2, gp.GenGrow, type_, r).Nodes()
		}, ps, r)
		fmt.Printf("After mut %d:\n", len(ret.Nodes()))
		fmt.Println(ret)
		printNodes(ret)
		fmt.Printf("mut and CX tree len(%d): %s=%s\n", len(ret.Nodes()), ret, ret.Compile())
		count--
	}
	fmt.Printf("Reached %d iteration\n", 100-count)
}

func printNodes(pt *gp.PrimitiveTree) {
	for _, n := range pt.Nodes() {
		fmt.Printf("%s ", n.Name())
	}
	fmt.Println("")
}

// func main() {
// 	var add gp.PrimitiveFunc = func(a ...interface{}) interface{} {
// 		ret := 0
// 		for _, n := range a {
// 			ret += n.(int)
// 		}
// 		return ret
// 	}
// 	var sub gp.PrimitiveFunc = func(a ...interface{}) interface{} {
// 		ret := a[0].(int)
// 		for _, n := range a[1:] {
// 			ret -= n.(int)
// 		}
// 		return ret
// 	}
// 	pset := gp.NewPrimitiveSet([]reflect.Kind{reflect.Int}, reflect.Int)
// 	pset.AddPrimitive(gp.NewPrimitive("add", add, []reflect.Kind{reflect.Int, reflect.Int}, reflect.Int))
// 	pset.AddPrimitive(gp.NewPrimitive("sub", sub, []reflect.Kind{reflect.Int, reflect.Int}, reflect.Int))
// 	pset.AddPrimitive(operator.mul, 2)
// 	pset.AddPrimitive(protectedDiv, 2)
// 	pset.AddPrimitive(operator.neg, 1)
// 	pset.AddPrimitive(math.cos, 1)
// 	pset.AddPrimitive(math.sin, 1)
// 	pset.addEphemeralConstant("rand101", partial(random.randint, -1, 1))
// }

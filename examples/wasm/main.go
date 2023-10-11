package main

import (
	"fmt"
	// "main/gp"
	"syscall/js"
)

func eaSimple() js.Func {
	jsonFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) != 1 {
			return "Invalid no of arguments passed"
		}
		inputJSON := args[0].String()
		fmt.Printf("input %s\n", inputJSON)
		return "hello"
	})
	return jsonFunc
}

func main() {
	fmt.Println("Go Web Assembly")
	js.Global().Set("EASimple", eaSimple())
	<-make(chan bool)
}

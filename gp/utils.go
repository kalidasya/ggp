package gp

import (
	"fmt"
	"strings"

	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
)

func Max(a int, b int) int {
	if a >= b {
		return a
	}
	return b
}

func Append(s []int, times int, value int) []int {
	for i := 0; i < times; i++ {
		s = append(s, value)
	}
	return s
}

func Pop[T any](s []T) ([]T, T) {
	return s[:len(s)-1], s[len(s)-1]
}

func replaceInRange(stack []Node, start, end int, insert ...Node) []Node {
	stack = slices.Clone(stack)
	fmt.Printf("Orig slice deleting from %d to %d:\n", start, end)
	PrintNodes(stack)
	stack = slices.Delete(stack, start, end)
	lenNeeded := len(stack) + len(insert)
	fmt.Println("After delete slice:")
	PrintNodes(stack)
	fmt.Println("inserting")
	PrintNodes(insert)
	fmt.Printf("at: %d\nstack len: %d cap: %d total needed: %d \n", start, len(stack), cap(stack), lenNeeded)
	if lenNeeded > cap(stack) {
		stack = slices.Grow(stack, lenNeeded-cap(stack))
		fmt.Printf("stack grown to len: %d cap: %d \n", len(stack), cap(stack))
	}
	stack = slices.Insert(stack, start, insert...)
	fmt.Printf("stack len after insert: %d cap: %d \n", len(stack), cap(stack))
	PrintNodes(stack)
	return stack
}

func Intersect[T constraints.Ordered](s1 []T, s2 []T) []T {
	// type1Keys := maps.Keys(types1)
	slices.Sort(s1)
	slices.Compact(s1)

	// type2Keys := maps.Keys(types2)
	slices.Sort(s2)
	slices.Compact(s2)

	var intersection []T
	for _, t1 := range s1 {
		for _, t2 := range s2 {
			if t1 == t2 {
				intersection = append(intersection, t1)
			}
		}
	}
	return intersection
}

func NodesAsString(nodes []Node) string {
	var b strings.Builder
	for _, n := range nodes {
		fmt.Fprintf(&b, "%s ", n.Name())
	}
	return b.String()
}

func PrintNodes(nodes []Node) {
	fmt.Println(NodesAsString(nodes))
}

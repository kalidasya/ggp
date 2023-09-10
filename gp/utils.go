package gp

import (
	"fmt"
	"reflect"
	"strings"

	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
)

type nodeString struct {
	node Node
	str  []string
}

type nodeInterface struct {
	node Node
	args []interface{}
}

type stackItem struct {
	i int
	t reflect.Kind
}

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

func ReplaceInRange[T any](stack []T, start, end int, insert ...T) []T {
	stack = slices.Clone(stack) // todo create unittest
	stack = slices.Delete(stack, start, end)
	stack = slices.Insert(stack, start, insert...)
	return stack
}

func Intersect[T constraints.Ordered](s1 []T, s2 []T) []T {
	slices.Sort(s1)
	slices.Compact(s1)
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

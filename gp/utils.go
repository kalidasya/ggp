package gp

import (
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
	"math/rand"
	"time"
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
	stack = slices.Delete(stack, start, end)
	stack = slices.Insert(stack, start, insert...)
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

type Random interface {
	Intn(int) int
	Float32() float32
}

type RealRandom struct{}

func (r *RealRandom) Intn(n int) int {
	return rand.New(rand.NewSource(time.Now().UnixNano())).Intn(n)
}

func (r *RealRandom) Float32() float32 {
	return rand.New(rand.NewSource(time.Now().UnixNano())).Float32()
}

func NewRealRandom() *RealRandom {
	return &RealRandom{}
}

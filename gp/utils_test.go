package gp

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type NodeI interface {
	Name() string
}

type TestNode struct {
	name string
}

func (tn *TestNode) Name() string {
	return tn.name
}

var _ NodeI = new(TestNode)

func TestReplaceInSlice(t *testing.T) {
	s := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	type TestCase struct {
		start    int
		end      int
		insert   []int
		expected []int
	}
	for _, tc := range []TestCase{
		{
			start:    0,
			end:      1,
			insert:   []int{11, 12},
			expected: []int{11, 12, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		},
		{
			start:    9,
			end:      10,
			insert:   []int{11, 12},
			expected: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 11, 12},
		},
		{
			start:    3,
			end:      4,
			insert:   []int{11, 12},
			expected: []int{1, 2, 3, 11, 12, 5, 6, 7, 8, 9, 10},
		},
		{
			start:    0,
			end:      3,
			insert:   []int{11, 12},
			expected: []int{11, 12, 4, 5, 6, 7, 8, 9, 10},
		},
		{
			start:    7,
			end:      10,
			insert:   []int{11, 12},
			expected: []int{1, 2, 3, 4, 5, 6, 7, 11, 12},
		},
		{
			start:    3,
			end:      6,
			insert:   []int{11, 12},
			expected: []int{1, 2, 3, 11, 12, 7, 8, 9, 10},
		},
		{
			start:    8,
			end:      10,
			insert:   []int{11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23},
			expected: []int{1, 2, 3, 4, 5, 6, 7, 8, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23},
		},
	} {
		t.Run(fmt.Sprintf("[%d:%d]<-%v", tc.start, tc.end, tc.insert), func(t *testing.T) {
			assert.Equal(t, tc.expected, ReplaceInRange(s, tc.start, tc.end, tc.insert...))
		})
	}
	node1 := &TestNode{name: "node1"}
	node2 := &TestNode{name: "node2"}
	node3 := &TestNode{name: "node3"}
	node4 := &TestNode{name: "node4"}
	nodesOriginal := []NodeI{
		node1, node2, node3, node4,
	}
	t.Run("Interface1", func(t *testing.T) {
		assert.Equal(t, []NodeI{node3, node3, node2, node3, node4}, ReplaceInRange(nodesOriginal, 0, 1, []NodeI{node3, node3}...))
	})
	t.Run("Interface2", func(t *testing.T) {
		assert.Equal(t, []NodeI{node1, node2, node3, node1, node1}, ReplaceInRange(nodesOriginal, 3, 4, []NodeI{node1, node1}...))
	})
}

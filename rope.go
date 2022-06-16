package rope

import (
	"strings"
)

const (
	maxDepth      = 64
	maxLeafSize   = 4096
	balanceFactor = 8
)

// A Rope is a data structure for storing long runs of text.
// Ropes are persistent: there is no way to modify an existing rope.
// Instead, all operations return a new rope with the requested changes.
//
// This persistence makes it easy to store old versions of a Rope just by holding on to old roots.
type Rope struct {
	content       string
	length, depth int
	left, right   *Rope
}

var empty = &Rope{}

// Return a new empty rope.
func New() *Rope {
	return empty
}

// Return a new rope with the contents of string s.
func NewString(s string) *Rope {
	return &Rope{content: s, length: len(s)}
}

// Return a new rope that is the concatenation of this rope and the other rope.
func (rope *Rope) Append(other *Rope) *Rope {
	switch {
	case rope.length == 0:
		return other
	case other.length == 0:
		return rope
	case rope.length+other.length <= maxLeafSize:
		return NewString(rope.String() + other.String())
	default:
		depth := rope.depth
		if other.depth > depth {
			depth = other.depth
		}
		return (&Rope{
			length: rope.length + other.length,
			depth:  depth + 1,
			left:   rope,
			right:  other,
		}).rebalanceIfNeeded()
	}
}

// Return a new rope that is the concatenation of this rope and string s.
func (rope *Rope) AppendString(other string) *Rope {
	return rope.Append(NewString(other))
}

// Return a new rope with length bytes at offset deleted.
func (rope *Rope) Delete(offset, length int) *Rope {
	if length == 0 || offset == rope.length {
		return rope
	}

	left, right := rope.Split(offset)
	_, newRight := right.Split(length)
	return left.Append(newRight)
}

// Returns true if this rope is equal to other.
func (rope *Rope) Equal(other *Rope) bool {
	if rope == other {
		return true
	}

	if rope.length != other.length {
		return false
	}

	for i := 0; i < rope.length; i++ {
		if rope.Index(i) != other.Index(i) {
			return false
		}
	}

	return true
}

// Return the byte at the given index.
func (rope *Rope) Index(at int) byte {
	if rope.isLeaf() {
		return rope.content[at]
	}

	if at < rope.left.length {
		return rope.left.Index(at)
	}

	return rope.right.Index(at - rope.left.length)
}

// Return a new rope with the contents of other inserted at the given index.
func (rope *Rope) Insert(at int, other *Rope) *Rope {
	if at == 0 {
		return other.Append(rope)
	}

	if at == rope.length {
		return rope.Append(other)
	}

	left, right := rope.Split(at)
	return left.Append(other).Append(right)
}

// Return a new rope with the contents of string other inserted at the given index.
func (rope *Rope) InsertString(at int, other string) *Rope {
	return rope.Insert(at, NewString(other))
}

// Return the length of the rope in bytes.
func (rope *Rope) Length() int {
	return rope.length
}

// Return a new version of this rope that is balanced for better performance.
// Generally speaking, this will be invoked automatically during the course of other operations and
// thus only needs to be called if you know you'll be generating a lot of unbalanced ropes.
func (rope *Rope) Rebalance() *Rope {
	if rope.isBalanced() {
		return rope
	}

	var leaves []*Rope
	rope.walk(func(node *Rope) {
		if node.isLeaf() {
			leaves = append(leaves, node)
		}
	})

	return merge(leaves, 0, len(leaves))
}

// Returns two new ropes, one containing the content to the left of the given index and the other the content to the right.
func (rope *Rope) Split(at int) (*Rope, *Rope) {
	switch {
	case rope.isLeaf():
		return NewString(rope.content[0:at]), NewString(rope.content[at:])

	case at == 0:
		return empty, rope

	case at == rope.length:
		return rope, empty

	case at < rope.left.length:
		left, right := rope.left.Split(at)
		return left, right.Append(rope.right)

	case at > rope.left.length:
		left, right := rope.right.Split(at - rope.left.length)
		return rope.left.Append(left), right

	default:
		return rope.left, rope.right
	}
}

// Return the contents of the rope as a string.
func (rope *Rope) String() string {
	if rope.isLeaf() {
		return rope.content
	}

	var builder strings.Builder
	rope.walk(func(node *Rope) {
		if node.isLeaf() {
			builder.WriteString(node.content)
		}
	})

	return builder.String()
}

func (rope *Rope) isBalanced() bool {
	switch {
	case rope.depth >= len(fibonacci)-2:
		return false
	case rope.isLeaf():
		return true
	default:
		return fibonacci[rope.depth+2] <= rope.length
	}
}

func (rope *Rope) isLeaf() bool {
	return rope.left == nil
}

func (rope *Rope) leafForOffset(at int) (*Rope, int) {
	if rope.isLeaf() {
		return rope, at
	}

	if at < rope.left.length {
		return rope.left.leafForOffset(at)
	}

	return rope.right.leafForOffset(at - rope.left.length)
}

func (rope *Rope) rebalanceIfNeeded() *Rope {
	if rope.isLeaf() || rope.isBalanced() || abs(rope.left.depth-rope.right.depth) > balanceFactor {
		return rope
	}

	return rope.Rebalance()
}

func (rope *Rope) walk(callback func(*Rope)) {
	if rope.isLeaf() {
		callback(rope)
	} else {
		rope.left.walk(callback)
		callback(rope)
		rope.right.walk(callback)
	}
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func merge(leaves []*Rope, start, end int) *Rope {
	length := end - start
	switch length {
	case 1:
		return leaves[start]
	case 2:
		return leaves[start].Append(leaves[start+1])
	default:
		mid := start + length/2
		return merge(leaves, start, mid).Append(merge(leaves, mid, end))
	}
}

var fibonacci []int

func init() {
	// The heurstic for whether a rope is balanced depends on the Fibonacci sequence;
	// we initialize the table of Fibonacci numbers here.
	first := 0
	second := 1

	for c := 0; c < maxDepth+3; c++ {
		next := 0
		if c <= 1 {
			next = c
		} else {
			next = first + second
			first = second
			second = next
		}
		fibonacci = append(fibonacci, next)
	}
}

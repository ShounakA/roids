package col

import "fmt"

// A Node type to hold "any" value
// Can see prev and next node
type Node[T any] struct {
	prev  *Node[T]
	next  *Node[T]
	value *T
}

// An interface representing the Stack type
type IStack[T any] interface {
	Push(newItem T)
	Pop() *T
	Display()
	String() string
	GetSize() uint64
	Reverse()
}

// A Stack type that can hold "any" one type of value
type Stack[T any] struct {
	head   *Node[T]
	length uint64
}

// Create a new stack collection
func NewStack[T any](initValue *T) *Stack[T] {
	head := new(Node[T])
	head.value = initValue
	head.next = new(Node[T])
	var length uint64
	if initValue == nil {
		length = 0
	} else {
		length = 1
	}
	return &Stack[T]{head: head, length: length}
}

func (stack *Stack[T]) Push(newItem T) {
	tempHead := stack.head
	stack.head = new(Node[T])
	stack.head.value = &newItem
	stack.head.next = tempHead
	stack.length++
}

func (stack *Stack[T]) Pop() *T {
	to_pop := stack.head
	stack.head = stack.head.next
	stack.length--
	return to_pop.value
}

func (stack *Stack[T]) Display() {
	tracker := stack.head
	for tracker.next != nil && !IsZero(tracker.value) {
		println(*tracker.value)
		tracker = tracker.next
	}
}

func (stack *Stack[T]) String() string {
	stack_display := "HEAD -> "
	tracker := stack.head
	for tracker.next != nil && !IsZero(tracker.value) {
		stack_display += fmt.Sprintf("%v -> ", *tracker.value)
		tracker = tracker.next
	}
	stack_display += "END"
	return stack_display
}

func (stack *Stack[T]) GetSize() uint64 {
	return stack.length
}

func Zero[T any]() T {
	return *new(T)
}

func IsZero[T comparable](v T) bool {
	return v == *new(T)
}

// Reverse reverses the order of the elements in the stack.
func (stack *Stack[T]) Reverse() {
	// If the stack has 0 or 1 elements, it is already reversed.
	if stack.length <= 1 {
		return
	}

	var zero T
	tempStack := NewStack[T](&zero)
	tempStack.Pop() // Make the stack truly empty.

	for stack.length > 0 {
		item := stack.Pop()
		if item != nil {
			tempStack.Push(*item)
		}
	}

	stack.head = tempStack.head
	stack.length = tempStack.length
}
